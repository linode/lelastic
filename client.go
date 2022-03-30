package main

import (
	"context"
	"fmt"
	"sync"

	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/pkg/server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Client is the client
type Client struct {
	c         *server.BgpServer
	ips       *[]IPNet
	ipv6Plen  int
	community string
	wg        *sync.WaitGroup
}

// NewClient instantiates a new client connection
func NewClient(c string, ips *[]IPNet) (*Client, error) {
	maxSize := 256 << 20
	grpcOpts := []grpc.ServerOption{grpc.MaxRecvMsgSize(maxSize), grpc.MaxSendMsgSize(maxSize)}

	cl := server.NewBgpServer(server.GrpcListenAddress(apiListen), server.GrpcOption(grpcOpts))

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go cl.Serve()

	if err := cl.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			As:         myAsn,
			RouterId:   id,
			ListenPort: -1,
		},
	}); err != nil {
		return nil, err
	}

	return &Client{
		c:         cl,
		ips:       ips,
		ipv6Plen:  64,
		community: c,
		wg:        wg,
	}, nil
}

// AddRs adds route-server peerings to the bgp session
func (c *Client) AddRs(rs string) error {
	n := &api.Peer{
		ApplyPolicy: &api.ApplyPolicy{
			ExportPolicy: &api.PolicyAssignment{
				Name:          "routeServer-out",
				Direction:     api.PolicyDirection_EXPORT,
				DefaultAction: api.RouteAction_ACCEPT,
			},
			ImportPolicy: &api.PolicyAssignment{
				Name:          "routeServer-in",
				Direction:     api.PolicyDirection_IMPORT,
				DefaultAction: api.RouteAction_REJECT,
			},
		},
		Conf: &api.PeerConf{
			NeighborAddress: rs,
			PeerAs:          rsAsn,
			Description:     "route server",
		},
		EbgpMultihop: &api.EbgpMultihop{
			Enabled:     true,
			MultihopTtl: 10,
		},
		Timers: &api.Timers{
			Config: &api.TimersConfig{
				ConnectRetry:      5,
				HoldTime:          9,
				KeepaliveInterval: 3,
			},
		},
		AfiSafis: []*api.AfiSafi{
			{
				Config: &api.AfiSafiConfig{
					Family: &api.Family{
						Afi:  api.Family_AFI_IP,
						Safi: api.Family_SAFI_UNICAST,
					},
					Enabled: true,
				},
			},
			{
				Config: &api.AfiSafiConfig{
					Family: &api.Family{
						Afi:  api.Family_AFI_IP6,
						Safi: api.Family_SAFI_UNICAST,
					},
					Enabled: true,
				},
			},
		},
	}

	if err := c.c.AddPeer(context.Background(), &api.AddPeerRequest{
		Peer: n,
	}); err != nil {
		return fmt.Errorf("failed adding neighbor: %w", err)
	}

	return nil
}

// AddStaticRoute adds a static route in gobgp incl some BGP attributes for export
func (c *Client) AddStaticRoute(nh string, p IPNet, cm string) error {
	path, err := getPath(p, nh, cm)
	if err != nil {
		return fmt.Errorf("unable to compile path pointer: %w", err)
	}

	_, err = c.c.AddPath(context.Background(), &api.AddPathRequest{
		Path: path,
	})
	if err != nil {
		return fmt.Errorf("failed adding route %v: %w", p, err)
	}

	return nil
}

// AddRoutes adds a static route for all IPs monitored
func (c *Client) AddRoutes() error {
	for _, ip := range *c.ips {
		if err := c.AddStaticRoute("", ip, c.community); err != nil {
			return err
		}
		log.WithFields(log.Fields{"Topic": "Route", "Route": ip, "Community": c.community}).Info("added route")
	}
	return nil
}
