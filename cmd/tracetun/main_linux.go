// Package main is a test utility: In order to test traceroute deterministically, this can be used to
// create a TUN listener which replies with ICMP.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/netip"
	"os"

	"golang.org/x/net/ipv4"
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

func reply(in []byte, config map[netip.Addr]parsedConfig) []byte {
	// Ignore non-IPv4
	if len(in) < 20 || in[0]>>4 != 4 {
		return nil
	}

	ttl := in[8]
	from := netip.AddrFrom4([4]byte(in[12:16]))
	dest := netip.AddrFrom4([4]byte(in[16:20]))

	destConfig, ok := config[dest]
	if !ok {
		return nil
	}

	if ttl > destConfig.DestinationTTL {
		return icmpMessage(dest, from, ipv4.ICMPTypeDestinationUnreachable, icmpPortUnreachable, in)
	}

	replySrc, ok := destConfig.hopsV4[int(ttl)]
	if !ok {
		return nil
	}

	return icmpMessage(replySrc, from, ipv4.ICMPTypeTimeExceeded, icmpTTLExceeded, in)
}

// Config defines the JSON structure of the config file.
type Config struct {
	// Destination IP address that traceroutes must be going to
	Destination string

	// DestinationTTL is the TTL to the destination.
	DestinationTTL byte

	// HopsV4 is a map of TTL to IPv4 address to reply from
	// Missing keys won't reply.
	HopsV4 map[int]string

	// TODO: HopsV6
}

type parsedConfig struct {
	Destination    netip.Addr
	DestinationTTL byte

	hopsV4 map[int]netip.Addr
}

func parseConfig(configFile string) (map[netip.Addr]parsedConfig, error) {
	cfgFile, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("opening config file: %v", err)
	}

	var cfgs []Config
	err = json.Unmarshal(cfgFile, &cfgs)
	if err != nil {
		log.Fatal(err)
	}

	config := make(map[netip.Addr]parsedConfig)

	for _, cfg := range cfgs {
		dest, err := netip.ParseAddr(cfg.Destination)
		if err != nil {
			return nil, fmt.Errorf("parsing traceroute destination address: %s", err)
		}

		hopsV4 := make(map[int]netip.Addr)
		for hop, ipv4hop := range cfg.HopsV4 {
			parsed, err := netip.ParseAddr(ipv4hop)
			if err != nil || !parsed.Is4() {
				return nil, fmt.Errorf("invalid ipv4 address: hop %d ip %s", hop, ipv4hop)
			}
			hopsV4[hop] = parsed
		}

		config[dest] = parsedConfig{
			Destination:    dest,
			DestinationTTL: cfg.DestinationTTL,
			hopsV4:         hopsV4,
		}
	}

	return config, nil
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: tracetun <interface> <config.json>")
	}

	config, err := parseConfig(os.Args[2])
	if err != nil {
		return
	}

	var debug = os.Getenv("DEBUG") == "1"

	if debug {
		log.Printf("Loaded configuration: %+v", config)
	}

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

		d := reply(buf[:n], config)
		if d != nil {
			if debug {
				log.Printf("replying to message: %x", buf[:n])
				log.Printf("               with: %x", d)
			}
			_, _ = tun.Write(d)
		} else if debug {
			log.Printf("not replying to: %x", buf[:n])
		}
	}
}
