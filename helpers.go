package main

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	api "github.com/osrg/gobgp/api"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
)

type IPNet net.IPNet

func (i IPNet) String() string {
	j, _ := i.Mask.Size()
	return fmt.Sprintf("%v/%v", i.IP, j)
}

func (i IPNet) Plen() uint32 {
	j, _ := i.Mask.Size()
	return uint32(j)
}

func IPNetFromAddr(a net.Addr) (*IPNet, error) {
	_, p, err := net.ParseCIDR(a.String())

	if err != nil {
		return nil, err
	}

	return &IPNet{
		IP:   p.IP,
		Mask: p.Mask,
	}, nil
}

// parse bgp community from string to uint32
func parseCommunity(c string) (uint32, error) {
	s := strings.SplitN(c, ":", 2)
	var i []uint32
	for _, o := range s {
		u, err := strconv.ParseUint(o, 10, 32)
		if err != nil {
			return 0, err
		}
		i = append(i, uint32(u))
	}
	parsed := i[0]*65536 + i[1]

	log.WithFields(log.Fields{"Topic": "parseCommunity", "Community": c}).Tracef("converted to: %v", parsed)
	return parsed, nil
}

// compile goBGP path type from prefix, next hop and bgp community
func getPath(p IPNet, nh string, myCom string) (*api.Path, error) {
	//convert human readable community to uint32
	c, err := parseCommunity(myCom)
	if err != nil {
		return nil, fmt.Errorf("failed to parse community: %w", err)
	}

	nhvar := []string{}
	if nh != "" {
		nhvar = append(nhvar, nh)
	}

	nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
		Prefix:    p.IP.String(),
		PrefixLen: p.Plen(),
	})

	var family *api.Family
	if p.IP.To4() == nil {
		family = &api.Family{
			Afi:  api.Family_AFI_IP6,
			Safi: api.Family_SAFI_UNICAST,
		}
	} else {
		family = &api.Family{
			Afi:  api.Family_AFI_IP,
			Safi: api.Family_SAFI_UNICAST,
		}
	}

	attrs, _ := ptypes.MarshalAny(&api.MpReachNLRIAttribute{
		Family:   family,
		NextHops: nhvar,
		Nlris:    []*any.Any{nlri},
	})

	com, _ := ptypes.MarshalAny(&api.CommunitiesAttribute{
		Communities: []uint32{c},
	})

	origin, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 2, // needs to be 2 for static route redistribution
	})

	log.WithFields(log.Fields{"Topic": "Helper", "Route": p}).Tracef("generated path NLRI %v with community %v", family.Afi, myCom)

	return &api.Path{
		Family: family,
		Nlri:   nlri,
		Pattrs: []*any.Any{origin, attrs, com},
	}, nil
}

func getIPs() ([]IPNet, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var ips []IPNet

	for _, addr := range addrs {
		ip, err := IPNetFromAddr(addr)
		if err != nil {
			log.WithFields(log.Fields{"Topic": "Helper", "Route": addr, "Error": "invalid IP"}).Warn("invalid IP")
		}
		if !ip.IP.IsGlobalUnicast() || ip.IP.IsPrivate() {
			log.WithFields(log.Fields{"Topic": "Helper", "Route": ip, "Error": "not acceptable elastic IP"}).Debug("not acceptable elastic IP")
			continue
		}
		if ip.Plen() != 32 && ip.Plen() != 56 && ip.Plen() != 64 {
			log.WithFields(log.Fields{"Topic": "Helper", "Route": ip, "Error": "not accepted prefix length"}).Debug("not accepted prefix length")
			continue
		}
		ips = append(ips, *ip)
		log.WithFields(log.Fields{"Topic": "Helper", "Route": ip}).Debug("handling prefix")
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("didn't find any configured elastic IPs")
	}

	return ips, nil
}
