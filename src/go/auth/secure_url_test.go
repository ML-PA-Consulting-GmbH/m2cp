package auth

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequireSecureURL(t *testing.T) {
	cases := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https", "https://api.test/oauth/token", false},
		{"https with port", "https://api.test:8443/token", false},
		{"http localhost", "http://localhost:8080/token", false},
		{"http 127.0.0.1", "http://127.0.0.1:54321/token", false},
		{"http ipv6 loopback", "http://[::1]:8080/token", false},
		{"http LOCALHOST case-insensitive", "http://LOCALHOST/token", false},
		{"http public rejected", "http://api.test/token", true},
		{"scheme-less rejected", "api.test/token", true},
		{"non-http(s) scheme rejected", "ftp://api.test/token", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := RequireSecureURL(tc.url, "test", false, nil)
			if tc.wantErr {
				assert.ErrorIs(t, err, ErrInsecureURL)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequireSecureURL_AllowInsecure(t *testing.T) {
	const insecure = "http://api.test/oauth/token" // non-loopback http

	t.Run("insecure without allow-insecure errors and does not warn", func(t *testing.T) {
		var warn bytes.Buffer
		err := RequireSecureURL(insecure, "token endpoint", false, &warn)
		assert.ErrorIs(t, err, ErrInsecureURL)
		assert.Empty(t, warn.String(), "no warning is written when the URL is rejected")
	})

	t.Run("insecure with allow-insecure warns and proceeds", func(t *testing.T) {
		var warn bytes.Buffer
		err := RequireSecureURL(insecure, "token endpoint", true, &warn)
		assert.NoError(t, err)
		assert.Contains(t, warn.String(), "--allow-insecure")
		assert.Contains(t, warn.String(), "token endpoint")
	})

	t.Run("secure URL never warns regardless of allow-insecure", func(t *testing.T) {
		var warn bytes.Buffer
		assert.NoError(t, RequireSecureURL("https://api.test/oauth/token", "token endpoint", true, &warn))
		assert.Empty(t, warn.String())
	})

	t.Run("loopback http is allowed without a warning", func(t *testing.T) {
		var warn bytes.Buffer
		assert.NoError(t, RequireSecureURL("http://localhost:8080/token", "token endpoint", false, &warn))
		assert.Empty(t, warn.String())
	})
}

func TestM2MLogin_RejectsInsecureTokenEndpoint(t *testing.T) {
	// store is https, but the (overridden) token endpoint is cleartext http.
	// Without --allow-insecure this must be rejected before any request is made.
	_, err := M2MLogin(t.Context(), M2MLoginInput{
		StoreURL:              "https://store.test/graphql",
		ClientID:              "c",
		OrgID:                 "o",
		Secret:                "s",
		TokenEndpointOverride: "http://idp.test/oauth/token", // non-loopback http
		AudienceOverride:      testAudience,
	})
	assert.ErrorIs(t, err, ErrInsecureURL)
}

// TestM2MLogin_AllowInsecureTokenEndpoint_Warns proves --allow-insecure is handed
// down to the token-endpoint check: the cleartext scheme no longer blocks the
// request (the login instead fails later at the network layer) and a warning is
// emitted to the provided writer.
func TestM2MLogin_AllowInsecureTokenEndpoint_Warns(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	var warn bytes.Buffer
	_, err := M2MLogin(ctx, M2MLoginInput{
		StoreURL:              "https://store.test/graphql",
		ClientID:              "c",
		OrgID:                 "o",
		Secret:                "s",
		TokenEndpointOverride: "http://idp.test/oauth/token", // non-loopback http, will not resolve
		AudienceOverride:      testAudience,
		AllowInsecure:         true,
		Warn:                  &warn,
	})
	assert.NotErrorIs(t, err, ErrInsecureURL, "the scheme check must be bypassed by --allow-insecure")
	assert.Contains(t, warn.String(), "--allow-insecure")
}
