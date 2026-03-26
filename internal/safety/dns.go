package safety

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"
)

// CheckDNS verifies that the target host can be resolved before starting the test.
func CheckDNS(targetURL string) error {
	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("target URL has no hostname")
	}

	// Skip DNS check for IP addresses
	if net.ParseIP(host) != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return fmt.Errorf("DNS resolution failed for %q: %w — check the hostname or your network", host, err)
	}

	if len(addrs) == 0 {
		return fmt.Errorf("DNS resolution returned no addresses for %q", host)
	}

	return nil
}
