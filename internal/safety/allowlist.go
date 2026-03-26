package safety

import (
	"fmt"
	"net/url"
	"slices"
)

// CheckHost verifies that the target URL's host is in the allowed hosts list.
func CheckHost(targetURL string, allowHosts []string) error {
	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL %q: %w", targetURL, err)
	}

	host := u.Host
	if host == "" {
		return fmt.Errorf("target URL %q has no host", targetURL)
	}

	if !slices.Contains(allowHosts, host) {
		return fmt.Errorf("host %q is not in allow_hosts %v — execution blocked", host, allowHosts)
	}

	return nil
}
