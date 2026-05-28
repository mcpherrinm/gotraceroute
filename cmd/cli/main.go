package main

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/mcpherrinm/gotraceroute/probe"
)

func main() {
	ipv4Flag := flag.Bool("4", false, "Use IPv4")
	ipv6Flag := flag.Bool("6", false, "Use IPv6")
	maxTTLFlag := flag.Int("m", 30, "Max TTL")
	port := flag.Int("P", 9000, "UDP target port")

	flag.Parse()

	narg := flag.NArg()
	if narg != 1 {
		fmt.Printf("Expected 1 destination, not %d\n", narg)
		return
	}

	destination := flag.Args()[0]

	fmt.Printf("ipv4=%v ipv6=%v max_ttl=%v port=%v destination=%v\n", *ipv4Flag, *ipv6Flag, *maxTTLFlag, *port, destination)

	ips, err := net.LookupIP(destination)
	if err != nil {
		fmt.Printf("failed looking up %q: %s", destination, err.Error())
		return
	}
	if len(ips) == 0 {
		fmt.Printf("failed looking up %q: no IPs", destination)
		return
	}

	for i := 1; i<*maxTTLFlag; i++ {
		fmt.Printf("%v %v %v\n", ips[0], *port, i)
		probe.Send(context.Background(), ips[0], *port, i)
	}
}
