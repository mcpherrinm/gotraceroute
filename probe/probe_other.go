//go:build !linux

package probe

import (
	"context"
	"fmt"
	"net"
)

// UDP Unsupported except on linux
func UDP(_ context.Context, _ net.IP, _ int, _ int) (Result, error) {
	return Result{}, fmt.Errorf("UDP traceroute not supported on this platform")
}
