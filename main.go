package main

import (
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
)

const (
	apiListen          = "127.0.0.1:50051"
	myAsn              = 65001
	rsAsn              = 65000
	id                 = "10.0.0.1"
	communityPrimary   = "65000:1"
	communitySecondary = "65000:2"
)

func main() {
	primary := flag.Bool("primary", false, "Advertise as primary (default if -secondary not specified)")
	secondary := flag.Bool("secondary", false, "Advertise as secondary")
	loglevel := flag.String("loglevel", "info", "Set log level: trace, debug, info or warn")
	logjson := flag.Bool("logjson", false, "Set log format to json")
	dcid := flag.Int("dcid", 0, "Your Linode data center id")
	send56 := flag.Bool("send56", false, "Advertise IPv6 as /56 subnet (defaults to /64)")
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
		log.WithFields(log.Fields{"Topic": "Main"}).Fatal("Required -dcid not provided")
	}

	if *primary && *secondary {
		flag.Usage()
		log.WithFields(log.Fields{"Topic": "Main"}).Fatal("Use either -primary or -secondary flag, not both!")
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
			Warn("Unknown log level, only trace, debug, info and warn are supported, falling back to loglevel info")
		log.SetLevel(log.InfoLevel)
	}

	myCommunity := communityPrimary;
	if *secondary {
		myCommunity = communitySecondary
	}

	//ips
	v6Mask := 64
	if *send56 {
		v6Mask = 56
	}

	ips, err := getIPs(v6Mask, *allIfs, flag.Args())
	if err != nil {
		log.WithFields(log.Fields{"Topic": "Main"}).Fatalf("No IPs: %v", err)
	}

	c, err := NewClient(myCommunity, ips)
	if err != nil {
		log.WithFields(log.Fields{"Topic": "Main"}).Fatal("Failed to initiate the BGP client: ", err)
	}

	c.wg.Add(1)

	for i := 1; i <= 4; i++ {
		rs := fmt.Sprintf("2600:3c0f:%d:34::%d", *dcid, i)
		if err := c.AddRs(rs); err != nil {
			log.WithFields(log.Fields{"Topic": "Neighbor", "Neighbor": rs}).Fatal("Failed adding neighbor")
		}
		// log.WithFields(log.Fields{"Topic": "Neighbor", "Neighbor": rs}).Info("added neighbor")
	}

	if err := c.AddRoutes(); err != nil {
		log.WithFields(log.Fields{"Topic": "IPs"}).Fatal("Failed adding IP advertisements: ", err)
	}

	log.WithFields(log.Fields{"Topic": "Main"}).Info("Running....")
	c.wg.Wait()
}
