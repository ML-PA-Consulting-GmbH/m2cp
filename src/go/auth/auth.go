package auth

import (
	"regexp"
)

// CheckAuthHeader Check if the auth header is valid.
// param auth_header: the authorization header sent by snapd, (includes root and discharge macaroon)
// return: the user account (e-mail) if verified, or empty string if not

func getFieldFromHeader(authHeader string, name string) string {
	r, _ := regexp.Compile(name + `\s*=\s*"*\s*([\w-]+)`)
	matches := r.FindStringSubmatch(authHeader)
	if len(matches) != 2 {
		return ""
	}
	raw := matches[1]
	return raw
}
