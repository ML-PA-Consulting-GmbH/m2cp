package coap

import (
	"fmt"
	"net/netip"
	"strings"
)

// ParseHostPort parses a host:port string and returns a netip.AddrPort.
// It handles various IPv6 formats including zone identifiers and missing ports.
// If no port is specified, it defaults to 5683.
func ParseHostPort(host string) (netip.AddrPort, error) {
	const defaultPort = 5683

	// Try direct parse first
	addrPort, err := netip.ParseAddrPort(host)
	if err == nil {
		return addrPort, nil
	}

	// Handle case where port is missing
	// Check if host is already bracketed IPv6 address
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		// Remove brackets and try to parse as address
		addrStr := host[1 : len(host)-1]
		addr, err := netip.ParseAddr(addrStr)
		if err != nil {
			return netip.AddrPort{}, fmt.Errorf("invalid address: %w", err)
		}
		return netip.AddrPortFrom(addr, defaultPort), nil
	}

	// Try parsing as plain address (IPv4 or IPv6 without brackets)
	addr, err := netip.ParseAddr(host)
	if err == nil {
		return netip.AddrPortFrom(addr, defaultPort), nil
	}

	return netip.AddrPort{}, fmt.Errorf("invalid host:port format: %s", host)
}
