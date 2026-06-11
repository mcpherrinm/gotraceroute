//go:build integration

package probe

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

// TestProbeIntegration assumes that tracetun is running with the in-repo configuration
func TestProbeIntegration(t *testing.T) {
	tests := []struct {
		ttl       int
		expectErr bool
	}{
		{ttl: 1, expectErr: false},
		{ttl: 2, expectErr: false},
		{ttl: 3, expectErr: false},
		{ttl: 4, expectErr: false},
		{ttl: 5, expectErr: false},
		{ttl: 6, expectErr: false},
		{ttl: 7, expectErr: false},
		{ttl: 8, expectErr: false},
		{ttl: 9, expectErr: true},
		{ttl: 10, expectErr: false},
		{ttl: 11, expectErr: false},
		{ttl: 12, expectErr: false},
		{ttl: 13, expectErr: false},
		{ttl: 14, expectErr: false},
		{ttl: 15, expectErr: false},
		{ttl: 16, expectErr: false},
	}
	destIP := net.IPv4(10, 100, 0, 15)
	for _, test := range tests {
		t.Run(fmt.Sprintf("ttl %d", test.ttl), func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond)
			defer cancel()

			result, err := UDP(ctx, destIP, 9000, test.ttl)
			if test.expectErr {
				if err == nil {
					t.Error("expected error, got none")
				}

				// Don't check the results if we expect an error
				return
			}
			if !test.expectErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if result.TTL != test.ttl {
				t.Errorf("expected TTL %d, got %d", test.ttl, result.Hop)
			}

			if result.RTT > time.Millisecond {
				t.Errorf("expected TTL under timeout %d, got %d", time.Millisecond, result.RTT)
			}
		})
	}
}
