package main

import (
	"flag"
	"fmt"
)

func main() {
	ipv4Flag := flag.Bool("4", false, "Use IPv4")
	ipv6Flag := flag.Bool("6", false, "Use IPv6")
	maxTTLFlag := flag.Int("m", 30, "Max TTL")

	flag.Parse()

	narg := flag.NArg()
	if narg != 1 {
		fmt.Printf("Expected 1 destination, not %d\n", narg)
		return
	}

	fmt.Printf("ipv4=%v ipv6=%v max_ttl=%v destination=%v\n", *ipv4Flag, *ipv6Flag, *maxTTLFlag, flag.Args()[0])
}
