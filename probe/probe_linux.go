package probe

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// UDP sends a single UDP probe, then waiting for its response
func UDP(ctx context.Context, to net.IP, port int, ttl int) (Result, error) {
	file, err := sock(ttl)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()

	sc, err := file.SyscallConn()
	if err != nil {
		return Result{}, err
	}

	sa, err := sockaddr(to, port)
	if err != nil {
		return Result{}, err
	}

	deadline, ok := ctx.Deadline()
	if ok {
		err = file.SetReadDeadline(deadline)
		if err != nil {
			return Result{}, err
		}
	}

	start := time.Now()

	var sendErr error
	err = sc.Control(func(fd uintptr) {
		sendErr = unix.Sendto(int(fd), []byte("ping"), 0, sa)
	})
	if err != nil {
		return Result{}, err
	}
	if sendErr != nil {
		return Result{}, sendErr
	}

	res := Result{
		To:  to,
		TTL: ttl,
	}

	var readErr error
	err = sc.Read(func(fd uintptr) bool {
		buf := make([]byte, 1500)
		oob := make([]byte, 1500)
		_, oobn, _, _, err := unix.Recvmsg(int(fd), buf, oob, unix.MSG_ERRQUEUE)
		if errors.Is(err, unix.EAGAIN) {
			return false
		}
		if err != nil {
			readErr = err
			return true
		}

		res.RTT = time.Since(start)

		hop, err := parse(oob[:oobn])
		if err != nil {
			readErr = err
			return true
		}

		res.Hop = hop
		return true
	})
	if err != nil {
		return Result{}, err
	}
	if readErr != nil {
		return Result{}, readErr
	}

	return res, nil
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
		return nil, fmt.Errorf("setting TTL: %w", err)
	}

	return os.NewFile(uintptr(fd), "probe"), nil
}

func parse(oob []byte) (net.IP, error) {
	messages, err := unix.ParseSocketControlMessage(oob)
	if err != nil {
		return nil, err
	}

	for _, message := range messages {
		if message.Header.Level != unix.IPPROTO_IP || message.Header.Type != unix.IP_RECVERR {
			// I'm not sure if we should get anything else, so log them.
			fmt.Printf("Got unexpected message len=%d level=%d type=%d\n%x\n", message.Header.Len, message.Header.Level, message.Header.Type, message.Data)
			continue
		}

		return parseSockExtendedErr(message.Data)
	}

	return nil, fmt.Errorf("no IP address found")
}

const sockExtendedErrSize = 16

// parseSockExtendedErr parses the entry from MSG_ERRQUEUE
// man 2 recv defines the structure
func parseSockExtendedErr(data []byte) (net.IP, error) {
	if len(data) < sockExtendedErrSize {
		return nil, fmt.Errorf("sockExtendedErr too short: length %d < expected %d", len(data), sockExtendedErrSize)
	}

	var see unix.SockExtendedErr

	rd := bytes.NewReader(data)
	err := binary.Read(rd, binary.NativeEndian, &see)
	if err != nil {
		return nil, err
	}

	switch see.Origin {
	case unix.SO_EE_ORIGIN_ICMP:
		var in unix.RawSockaddrInet4
		err = binary.Read(rd, binary.NativeEndian, &in)
		if err != nil {
			return nil, err
		}
		return net.IPv4(in.Addr[0], in.Addr[1], in.Addr[2], in.Addr[3]), nil
	case unix.SO_EE_ORIGIN_ICMP6:
		var in6 unix.RawSockaddrInet6
		err = binary.Read(rd, binary.NativeEndian, &in6)
		if err != nil {
			return nil, err
		}
		ip := make(net.IP, net.IPv6len)
		copy(ip, in6.Addr[:])
		return ip, nil
	default:
		return nil, fmt.Errorf("unknown origin %d", see.Origin)
	}
}
