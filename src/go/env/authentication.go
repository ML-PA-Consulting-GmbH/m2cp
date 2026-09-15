package env

import (
	"fmt"
	"strings"
)

type AuthenticationMethod int

const (
	SshAuthentication AuthenticationMethod = iota
	BrowserAuthentication
	M2MAuthentication
)

var allowedAuthenticationMethods = [...]AuthenticationMethod{
	SshAuthentication,
	BrowserAuthentication,
	M2MAuthentication,
}

func (method AuthenticationMethod) String() string {
	switch method {
	case SshAuthentication:
		return "ssh"
	case BrowserAuthentication:
		return "browser"
	case M2MAuthentication:
		return "m2m"
	default:
		return "unknown"
	}
}

func ParseAuthenticationMethod(method string) (AuthenticationMethod, error) {
	sanitizedMethod := strings.ToLower(method)
	for _, m := range allowedAuthenticationMethods {
		if m.String() == sanitizedMethod {
			return m, nil
		}
	}
	return BrowserAuthentication, fmt.Errorf("invalid authentication method: %q", method)
}

func IsValidAuthenticationMethod(method string) bool {
	_, err := ParseAuthenticationMethod(method)
	if err != nil {
		return false
	}
	return true
}

func ListingOfKnownAuthenticationMethods() string {
	methods := ""
	for idx, m := range allowedAuthenticationMethods {
		methods += fmt.Sprintf("\"%s\"", m.String())
		if idx < len(allowedAuthenticationMethods)-1 {
			methods += ", "
		}
	}
	return methods
}
