package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Reserved .test domains (RFC 6761) stand in for a real IdP so the unit tests
// never depend on a live tenant. They mirror the shape of the gateway's
// /auth/oidc-discovery response and may back a full httptest gateway later.
const (
	testIssuer            = "https://issuer.test/"
	testAudience          = "https://api.test"
	testTokenEndpoint     = "https://issuer.test/oauth/token"
	testAuthorizeEndpoint = "https://issuer.test/authorize"
	testLogoutEndpoint    = "https://issuer.test/logout"
)

// discoveryServer starts a test gateway serving /auth/oidc-discovery with the
// given handler and returns a matching --store URL (…/graphql), from which the
// discovery base URL is derived exactly as the browser flow does.
func discoveryServer(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/oidc-discovery", handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL + "/graphql"
}

func TestFetchOidcDiscovery_Valid(t *testing.T) {
	store := discoveryServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"providerName": "AUTH0",
			"issuers": ["` + testIssuer + `"],
			"audiences": ["` + testAudience + `"],
			"tokenEndpoint": "` + testTokenEndpoint + `",
			"authorizeEndpoint": "` + testAuthorizeEndpoint + `",
			"logoutEndpoint": "` + testLogoutEndpoint + `"
		}`))
	})

	disc, err := fetchOidcDiscovery(t.Context(), store)
	require.NoError(t, err)
	assert.Equal(t, testTokenEndpoint, disc.TokenEndpoint)
	assert.Equal(t, []string{testIssuer}, disc.Issuers)
	assert.Equal(t, []string{testAudience}, disc.Audiences)
	assert.Equal(t, "AUTH0", disc.ProviderName)
}

func TestFetchOidcDiscovery_NotFound_IsUnsupported(t *testing.T) {
	store := discoveryServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := fetchOidcDiscovery(t.Context(), store)
	assert.ErrorIs(t, err, ErrDiscoveryUnsupported)
}

func TestFetchOidcDiscovery_ServiceUnavailable(t *testing.T) {
	store := discoveryServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	_, err := fetchOidcDiscovery(t.Context(), store)
	assert.ErrorIs(t, err, ErrDiscoveryUnavailable)
	assert.NotErrorIs(t, err, ErrDiscoveryUnsupported)
}

// A proxy/ingress answering an unknown route with 200 + HTML must be treated as
// "unsupported", not mistaken for a valid discovery document.
func TestFetchOidcDiscovery_HtmlBody_IsUnsupported(t *testing.T) {
	store := discoveryServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<!doctype html><html><body>login</body></html>"))
	})
	_, err := fetchOidcDiscovery(t.Context(), store)
	assert.ErrorIs(t, err, ErrDiscoveryUnsupported)
}

func TestFetchOidcDiscovery_MissingTokenEndpoint_IsUnsupported(t *testing.T) {
	store := discoveryServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"providerName":"AUTH0","audiences":["` + testAudience + `"]}`))
	})
	_, err := fetchOidcDiscovery(t.Context(), store)
	assert.ErrorIs(t, err, ErrDiscoveryUnsupported)
}

func TestResolveTokenEndpoint(t *testing.T) {
	disc := &OidcDiscovery{TokenEndpoint: testTokenEndpoint}

	got, err := resolveTokenEndpoint("https://override.test/token", disc)
	require.NoError(t, err)
	assert.Equal(t, "https://override.test/token", got, "override wins")

	got, err = resolveTokenEndpoint("", disc)
	require.NoError(t, err)
	assert.Equal(t, testTokenEndpoint, got)

	_, err = resolveTokenEndpoint("", &OidcDiscovery{})
	assert.ErrorIs(t, err, ErrNoTokenEndpoint)
}

func TestResolveAudience(t *testing.T) {
	// override always wins
	got, err := resolveAudience("https://override.test", &OidcDiscovery{Audiences: []string{"https://a.test", "https://b.test"}})
	require.NoError(t, err)
	assert.Equal(t, "https://override.test", got)

	// exactly one discovered audience is used
	got, err = resolveAudience("", &OidcDiscovery{Audiences: []string{testAudience}})
	require.NoError(t, err)
	assert.Equal(t, testAudience, got)

	// multiple audiences without an override is ambiguous
	_, err = resolveAudience("", &OidcDiscovery{Audiences: []string{"https://a.test", "https://b.test"}})
	assert.ErrorIs(t, err, ErrMultipleAudiences)

	// none available
	_, err = resolveAudience("", &OidcDiscovery{})
	assert.ErrorIs(t, err, ErrNoAudience)
}
