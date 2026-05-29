// Package main is a traceroute CLI
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"slices"
	"time"

	"github.com/mcpherrinm/gotraceroute/probe"
)

func main() {
	ipv4Flag := flag.Bool("4", false, "Use IPv4")
	ipv6Flag := flag.Bool("6", false, "Use IPv6")
	maxTTLFlag := flag.Int("m", 30, "Max TTL")
	port := flag.Int("p", 9000, "UDP target port")

	flag.Parse()

	narg := flag.NArg()
	if narg != 1 {
		fmt.Printf("Expected 1 destination, not %d\n", narg)
		return
	}

	destination := flag.Args()[0]

	if *ipv4Flag && *ipv6Flag {
		fmt.Printf("Both -4 and -6 specified: Pick one")
		return
	}

	ips, err := net.LookupIP(destination)
	if err != nil {
		fmt.Printf("failed looking up %q: %s\n", destination, err.Error())
		return
	}
	if len(ips) == 0 {
		fmt.Printf("failed looking up %q: no IPs\n", destination)
		return
	}

	var ip net.IP
	if *ipv4Flag {
		idx := slices.IndexFunc(ips, func(ip net.IP) bool { return ip.To4() != nil })
		if idx == -1 {
			fmt.Printf("failed looking up %q: no IPv4 address\n", destination)
			return
		}
		ip = ips[idx]
	}

	if *ipv6Flag {
		idx := slices.IndexFunc(ips, func(ip net.IP) bool { return ip.To4() == nil })
		if idx == -1 {
			fmt.Printf("failed looking up %q: no IPv6 address\n", destination)
			return
		}
		ip = ips[idx]
	}
	if !*ipv4Flag && !*ipv6Flag {
		ip = ips[0]
	}

	for i := 1; i <= *maxTTLFlag; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		hop, err := probe.UDP(ctx, ip, *port, i)
		if err != nil {
			fmt.Printf("failed probe: %s\n", err.Error())
			continue
		}
		var name string
		names, err := net.LookupAddr(hop.Hop.String())
		if err != nil || len(names) == 0 {
			name = hop.Hop.String()
		} else {
			name = names[0]
		}

		fmt.Printf("%3d %s (%s) %s\n", i, name, hop.Hop, hop.RTT)

		if hop.Hop.Equal(ip) {
			return
		}
	}
}
