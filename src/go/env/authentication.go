package env

import (
	"fmt"
	"slices"
	"strings"
)

// AuthenticationMethod is the login mechanism. Each constant is its own
// canonical name — the value written to state.json and accepted by --method
type AuthenticationMethod string

const (
	SshAuthentication     AuthenticationMethod = "ssh"
	BrowserAuthentication AuthenticationMethod = "browser"
	M2MAuthentication     AuthenticationMethod = "m2m"
)

var allowedAuthenticationMethods = []AuthenticationMethod{
	SshAuthentication,
	BrowserAuthentication,
	M2MAuthentication,
}

// ParseAuthenticationMethod maps a (case-insensitive) --method value to a known AuthenticationMethod.
// On an unknown value it returns BrowserAuthentication — the default method — alongside the error,
// so a caller that ignores the error still can fall back to the default.
func ParseAuthenticationMethod(method string) (AuthenticationMethod, error) {
	candidate := AuthenticationMethod(strings.ToLower(method))
	if slices.Contains(allowedAuthenticationMethods, candidate) {
		return candidate, nil
	}
	return BrowserAuthentication, fmt.Errorf("invalid authentication method: %q", method)
}

func IsValidAuthenticationMethod(method string) bool {
	_, err := ParseAuthenticationMethod(method)
	return err == nil
}

// ListingOfKnownAuthenticationMethods returns the known methods as a quoted,
// comma-separated list (e.g. `"ssh", "browser", "m2m"`) for help and errors.
func ListingOfKnownAuthenticationMethods() string {
	quoted := make([]string, len(allowedAuthenticationMethods))
	for i, m := range allowedAuthenticationMethods {
		quoted[i] = fmt.Sprintf("%q", string(m))
	}
	return strings.Join(quoted, ", ")
}
