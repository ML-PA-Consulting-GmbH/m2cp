package auth

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

// ErrInsecureURL indicates an OAuth endpoint URL is not https and is not a
// loopback address. Loopback http is permitted only for local development and
// testing.
var ErrInsecureURL = errors.New("must use https")

// RequireSecureURL enforces that rawURL uses https, or uses http against a
// loopback host (localhost, 127.0.0.0/8, ::1) — loopback http permitted only for
// local development and testing. Sending client credentials or bearer tokens
// over cleartext http would expose them to a network attacker (RFC 6749 §10.8,
// RFC 9700 §2.6).
//
// A non-loopback http URL is rejected with an error wrapping ErrInsecureURL,
// unless allowInsecure is true, in which case the rejection is downgraded to a
// warning written to the io.Writer warn (os.Stderr when nil) and nil is returned,
// letting the caller proceed over cleartext.
// This --allow-insecure escape hatch is intended for local development only (e.g. docker container IPs).
// It is not recommended for production use, even inside a private network, because it undermines defense-in-depth
// and can expose client secrets and tokens to a network attacker.
//
// A malformed URL is always an error, and a secure URL never warns.
func RequireSecureURL(rawURL, label string, allowInsecure bool, warn io.Writer) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid %s URL %q: %w", label, rawURL, err)
	}
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if isLoopbackHost(u.Hostname()) {
			return nil
		}
	}

	insecureErr := fmt.Errorf("%s URL %q %w (cleartext http is allowed only for loopback)", label, rawURL, ErrInsecureURL)
	if !allowInsecure {
		return insecureErr
	}
	if warn == nil {
		warn = os.Stderr
	}
	fmt.Fprintf(warn, "warning: %s; continuing anyway because --allow-insecure was set. Credentials and tokens may be exposed on the network path — use only for local development.\n", insecureErr)
	return nil
}

// isLoopbackHost reports whether host is "localhost" or a loopback IP literal.
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
