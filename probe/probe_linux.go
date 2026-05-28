package probe

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// Result of a probe
type Result struct {
	To  net.IP
	Hop net.IP
	TTL int
	RTT time.Duration
}

func Send(ctx context.Context, to net.IP, port int, ttl int) (Result, error) {
	file, err := sock(ttl)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()

	sa, err := sockaddr(to, port)
	if err != nil {
		return Result{}, err
	}

	err = unix.Sendto(int(file.Fd()), []byte("ping"), 0, sa)
	if err != nil {
		return Result{}, err
	}

	return Result{}, fmt.Errorf("todo: %v", sa)
}

// sockaddr returns a unix.Sockaddr for (to, port)
// TODO: Only supports IPv4 right now
func sockaddr(to net.IP, port int) (unix.Sockaddr, error) {
	to4 := to.To4()
	if to4 == nil {
		return nil, fmt.Errorf("not an IPv4 address: %v", to)
	}
	addr := unix.SockaddrInet4{
		Port: port,
	}
	copy(addr.Addr[:], to4)

	return &addr, nil
}

// sock returns an os.File set up to use with the ttl
func sock(ttl int) (*os.File, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK, unix.IPPROTO_UDP)
	if err != nil {
		return nil, fmt.Errorf("probe socket: %w", err)
	}

	// Set IP_RECVERR so we get back our TTL errors
	err = unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_RECVERR, 1)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("setting IP_RECVERR: %w", err)
	}

	// Set the TTL
	err = unix.SetsockoptInt(fd, unix.IPPROTO_IP, unix.IP_TTL, ttl)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("Setting TTL: %w", err)
	}

	return os.NewFile(uintptr(fd), "probe"), nil
}
