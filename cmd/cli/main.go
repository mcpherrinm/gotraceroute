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

	ip, err := getIP(*ipv4Flag, *ipv6Flag, flag.Args()[0])
	if err != nil {
		fmt.Println(err)
		return
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

func getIP(ipv4 bool, ipv6 bool, destination string) (net.IP, error) {
	if ipv4 && ipv6 {
		return nil, fmt.Errorf("both -4 and -6 specified")
	}

	ips, err := net.LookupIP(destination)
	if err != nil {
		return nil, fmt.Errorf("failed looking up %q: %w", destination, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("failed looking up %q: no IPs", destination)
	}

	var ip net.IP
	if ipv4 {
		idx := slices.IndexFunc(ips, func(ip net.IP) bool { return ip.To4() != nil })
		if idx == -1 {
			return nil, fmt.Errorf("failed looking up %q: no IPv4 address", destination)
		}
		ip = ips[idx]
	}

	if ipv6 {
		idx := slices.IndexFunc(ips, func(ip net.IP) bool { return ip.To4() == nil })
		if idx == -1 {
			return nil, fmt.Errorf("failed looking up %q: no IPv6 address", destination)
		}
		ip = ips[idx]
	}

	if !ipv4 && !ipv6 {
		ip = ips[0]
	}

	return ip, nil
}
