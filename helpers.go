package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"bits.linode.com/netops/lelastic/pkg/helper"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	api "github.com/osrg/gobgp/v3/api"
	log "github.com/sirupsen/logrus"
)

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
func getPath(p helper.IPNet, nh string, myCom string) (*api.Path, error) {
	// convert human readable community to uint32
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

	log.WithFields(log.Fields{"Topic": "Helper", "Route": p}).
		Tracef("generated path NLRI %v with community %v", family.Afi, myCom)

	return &api.Path{
		Family: family,
		Nlri:   nlri,
		Pattrs: []*any.Any{origin, attrs, com},
	}, nil
}

// get all local IPs elegible to be elastic IP
func getIPs(v6Mask int, allIfs bool) (*[]helper.IPNet, error) {
	var addrs []net.Addr
	var err error

	if allIfs {
		addrs, err = net.InterfaceAddrs()
	} else {
		lo, err := net.InterfaceByName("lo")
		if err != nil {
			return nil, err
		}
		addrs, err = lo.Addrs()
	}
	if err != nil {
		return nil, err
	}

	sendMask := net.CIDRMask(v6Mask, 128)

	ips := make(map[string]*helper.IPNet)

	for _, addr := range addrs {
		ip, err := helper.IPNetFromAddr(addr)
		if err != nil {
			log.WithFields(log.Fields{"Topic": "Helper", "Route": addr, "Error": "invalid IP"}).Warn("invalid IP")
			continue
		}

		// ignore loopback IPs
		if ip.IP.IsLoopback() {
			log.WithFields(log.Fields{"Topic": "Helper", "Route": ip, "Warn": "not acceptable elastic IP"}).
				Trace("ignoring loopback IPs")
			continue
		}

		// ignore link local IPs
		if ip.IP.IsLinkLocalUnicast() {
			log.WithFields(log.Fields{"Topic": "Helper", "Route": ip, "Warn": "not acceptable elastic IP"}).
				Trace("ignoring linklocal IPs")
			continue
		}

		// for ipv4 only a /32 is acceptable
		if ip.IP.To4() != nil && ip.Plen() != 32 {
			log.WithFields(log.Fields{"Topic": "Helper", "Route": ip, "Warn": "not accepted prefix length"}).
				Warn("not accepted prefix length")
			continue
		}

		// for ipv6 lets find the greater subnet we're part of, make it a /64 (or if asked a /56) and advertise that
		if ip.IP.To4() == nil {
			if ip.Plen() != 64 && ip.Plen() != 56 {
				log.WithFields(log.Fields{"Topic": "Helper", "Route": ip, "Warn": "fixing prefix length"}).
					Warnf("fixing prefix lenth length to /%d", v6Mask)
				ip.Mask = sendMask
			}
			_, ipNew, err := net.ParseCIDR(ip.String())
			if err != nil {
				log.WithFields(log.Fields{"Topic": "Helper", "Route": ip, "Error": "invalid IP"}).
					Warnf("unable to supernet")
				continue
			}
			ip = &helper.IPNet{
				IP:   ipNew.IP,
				Mask: ipNew.Mask,
			}
		}

		ips[ip.String()] = ip
		log.WithFields(log.Fields{"Topic": "Helper", "Route": ip}).Debug("handling prefix")
	}

	var uniqIPs []helper.IPNet
	for _, ip := range ips {
		uniqIPs = append(uniqIPs, *ip)
	}

	if len(uniqIPs) == 0 {
		return nil, fmt.Errorf("didn't find any configured elastic IPs")
	}

	return &uniqIPs, nil
}
