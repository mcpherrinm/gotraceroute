// Package main is a test utility: In order to test traceroute deterministically, this can be used to
// create a TUN listener which replies with ICMP.
package main

import (
	"encoding/json"
	"log"
	"net/netip"
	"os"

	"golang.org/x/sys/unix"
)

// See https://www.kernel.org/doc/Documentation/networking/tuntap.txt
func tun(netName string) (*os.File, error) {
	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}

	ifreq, err := unix.NewIfreq(netName)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}

	// tun device with no packet information
	ifreq.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI)

	err = unix.IoctlIfreq(fd, unix.TUNSETIFF, ifreq)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}

	return os.NewFile(uintptr(fd), "/dev/net/tun"), nil
}

func reply(in []byte, hops map[netip.Addr]map[int]netip.Addr) []byte {
	// Ignore non-IPv4
	if len(in) < 20 || in[0]>>4 != 4 {
		return nil
	}

	ttl := in[8]
	from := netip.AddrFrom4([4]byte(in[12:16]))
	dest := netip.AddrFrom4([4]byte(in[16:20]))

	route, ok := hops[dest]
	if !ok {
		return nil
	}

	replySrc, ok := route[int(ttl)]
	if !ok {
		return nil
	}

	return icmpTimeExceeded(replySrc, from, in)
}

// Config defines the JSON structure of the config file.
type Config struct {
	// Destination IP address that traceroutes must be going to
	Destination string

	// HopsV4 is a map of TTL to IPv4 address to reply from
	// Missing keys won't reply.
	HopsV4 map[int]string

	// TODO: HopsV6
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: tracetun <interface> <config.json>")
	}

	cfgFile, err := os.ReadFile(os.Args[2])
	if err != nil {
		log.Fatalf("opening config file: %v", err)
	}

	cfg := Config{}
	err = json.Unmarshal(cfgFile, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	dest, err := netip.ParseAddr(cfg.Destination)
	if err != nil {
		log.Fatalf("parsing traceroute destination address: %s", err)
	}

	hopsV4 := make(map[netip.Addr]map[int]netip.Addr)
	hopsV4[dest] = make(map[int]netip.Addr)
	for hop, ipv4 := range cfg.HopsV4 {
		parsed, err := netip.ParseAddr(ipv4)
		if err != nil || !parsed.Is4() {
			log.Fatalf("invalid ipv4 address: hop %d ip %s", hop, ipv4)
		}
		hopsV4[dest][hop] = parsed
	}

	log.Printf("loaded hops: %v", hopsV4)

	tun, err := tun(os.Args[1])
	if err != nil {
		log.Fatalf("opening tun device: %v", err)
	}

	buf := make([]byte, 65535)
	for {
		n, err := tun.Read(buf)
		if err != nil {
			continue
		}

		log.Printf("got a message: %x", buf[:n])
		d := reply(buf[:n], hopsV4)
		if d != nil {
			log.Printf("replying %x", d)
			_, _ = tun.Write(d)
		}
	}
}
