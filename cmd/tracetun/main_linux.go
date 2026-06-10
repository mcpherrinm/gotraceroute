// Package main is a test utility: In order to test traceroute deterministically, this can be used to
// create a TUN listener which replies with ICMP.
package main

import (
	"log"
	"net"
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

func reply(in []byte, base net.IP) []byte {
	// Ignore non-IPv4
	if len(in) < 20 || in[0]>>4 != 4 {
		return nil
	}

	ttl := in[8]
	from := in[12:16]

	// TODO do hop offsets better
	replySrc := base.To4()
	replySrc[3] += ttl

	return icmpTimeExceeded(replySrc, from, in)
}

func main() {
	tun, err := tun("tun0")
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 65535)
	for {
		n, err := tun.Read(buf)
		if err != nil {
			continue
		}

		log.Printf("got a message: %x", buf[:n])
		d := reply(buf[:n], net.IPv4(162, 252, 205, 130))
		if d != nil {
			log.Printf("replying %x", d)
			_, _ := tun.Write(d)
		}
	}
}
