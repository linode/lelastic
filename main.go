package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
)

const (
	apiListen          = "127.0.0.1:50051"
	myAsn              = 65001
	rsAsn              = 65000
	id                 = "10.0.0.1"
	communityPrimary   = "65000:1"
	communitySecondary = "65000:2"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", []string(*i))
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var primaryIps arrayFlags
var secondaryIps arrayFlags

func main() {
	flag.Var(&primaryIps, "primary", "Advertise as primary for a specific IP. Must contain CIDR notation")
	flag.Var(&secondaryIps, "secondary", "Advertise as secondary for a specific IP. Must contain CIDR notation")
	loglevel := flag.String("loglevel", "info", "set log level: trace, debug, info or warn")
	logjson := flag.Bool("logjson", false, "set log format to json")
	dcid := flag.Int("dcid", 0, "dcid for your DC")
	send56 := flag.Bool("send56", false, "Advertise ipv6 as /56 subnet (defaults to /64)")
	allIfs := flag.Bool(
		"allifs",
		false,
		"Consider all interfaces when detecting elastic IP candidates (not just loopback)",
	)

	flag.Parse()

	if *logjson {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
			PadLevelText:  true,
			DisableColors: false,
		})
	}

	if *dcid <= 1 {
		flag.Usage()
		log.WithFields(log.Fields{"Topic": "Main"}).Fatal("dcid not provided, I need this info")
	}

	if primaryIps == nil && secondaryIps == nil {
		flag.Usage()
		log.WithFields(log.Fields{"Topic": "Main"}).Fatal("primary and/or secondary must be provided")
	}

	for _, ip := range primaryIps {
		if _, _, err := net.ParseCIDR(ip); err != nil {
			log.WithFields(log.Fields{"Topic": "Main"}).Fatalf("invalid primary ip: %s. Must be in CIDR notation", ip)
		}
	}

	for _, ip := range secondaryIps {
		if _, _, err := net.ParseCIDR(ip); err != nil {
			log.WithFields(log.Fields{"Topic": "Main"}).Fatalf("invalid secondary ip: %s. Must be in CIDR notation", ip)
		}
	}

	switch *loglevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.WithFields(log.Fields{"Topic": "Main"}).
			Warn("unknown log level, only trace, debug, info and warn are supported, falling back to loglevel info")
		log.SetLevel(log.InfoLevel)
	}

	var communityMap map[string]string
	communityMap = make(map[string]string)

	if len(primaryIps) > 0 {
		for _, ip := range primaryIps {
			communityMap[ip] = communityPrimary
		}
	}

	if len(secondaryIps) > 0 {
		for _, ip := range secondaryIps {
			communityMap[ip] = communitySecondary
		}
	}

	v6Mask := 64
	if *send56 {
		v6Mask = 56
	}

	allIps, err := getIPs(v6Mask, *allIfs)
	if err != nil {
		log.WithFields(log.Fields{"Topic": "Main"}).Fatalf("unable to detect IPs: %v", err)
	}

	var ips []IPNet

	// Filter the list of IPs to only include the ones we want to advertise
	if len(communityMap) > 0 {
		for _, ipData := range *allIps {
			ipString := ipData.String()
			if _, ok := communityMap[ipString]; ok {
				log.WithFields(log.Fields{"Topic": "Main"}).Infof("advertising IP %s with community %s", ipString, communityMap[ipString])
				ipData.community = communityMap[ipString]
				ips = append(ips, ipData)
			} else {
				log.WithFields(log.Fields{"Topic": "Main", "IP": ipString}).Warnf("not advetising IP %s as it was not specified", ipString)
			}
		}
	} else {
		log.WithFields(log.Fields{"Topic": "Main"}).Info("no IPs specified, advertising all IPs")
		ips = *allIps
	}

	if len(ips) != len(communityMap) {
		log.WithFields(log.Fields{
			"Topic":                 "Main",
			"Detected":              ips,
			"Requested Primaries":   primaryIps,
			"Requested Secondaries": secondaryIps,
		}).Fatal("Unable to detect all IPs specified. Check the IP addresses assigned to 'lo' or alternatively try the -allifs flag")
	}

	c, err := NewClient(&ips)
	if err != nil {
		log.WithFields(log.Fields{"Topic": "Main"}).Fatal("failed to initiate the client: ", err)
	}

	c.wg.Add(1)

	for i := 1; i <= 4; i++ {
		rs := fmt.Sprintf("2600:3c0f:%d:34::%d", *dcid, i)
		if err := c.AddRs(rs); err != nil {
			log.WithFields(log.Fields{"Topic": "Neighbor", "Neighbor": rs}).Fatal("failed adding neighbor")
		}
		// log.WithFields(log.Fields{"Topic": "Neighbor", "Neighbor": rs}).Info("added neighbor")
	}

	if err := c.AddRoutes(); err != nil {
		log.WithFields(log.Fields{"Topic": "IPs"}).Fatal("failed adding IP advertisements: ", err)
	}

	log.WithFields(log.Fields{"Topic": "Main"}).Info("Running....")
	c.wg.Wait()
}
