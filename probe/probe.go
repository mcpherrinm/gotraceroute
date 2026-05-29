// Package probe implements the platform-specific logic of sending out network
// requests and getting their responses.
package probe

import (
	"net"
	"time"
)

// Result of a probe
type Result struct {
	To  net.IP
	Hop net.IP
	TTL int
	RTT time.Duration
}
