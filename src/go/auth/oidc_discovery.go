package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OidcDiscovery is the machine-login discovery document served by the gateway at
// GET {base}auth/oidc-discovery. Field names match the gateway's camelCase JSON
// (Go's decoder also matches case-insensitively).
//
// Issuers and Audiences are arrays: the gateway returns the identity provider's full ValidIssuers/ValidAudiences.
// There is not yet a dedicated machine audience, so callers disambiguate with --audience when more than one is
// listed (see resolveAudience).
type OidcDiscovery struct {
	ProviderName      string   `json:"providerName"`
	Issuers           []string `json:"issuers"`
	Audiences         []string `json:"audiences"`
	TokenEndpoint     string   `json:"tokenEndpoint"`
	AuthorizeEndpoint string   `json:"authorizeEndpoint"`
	LogoutEndpoint    string   `json:"logoutEndpoint"`
}

const discoveryTimeout = 15 * time.Second

// Sentinel errors for machine-login discovery and endpoint/audience resolution.
// Callers match these with errors.Is.
var (
	// ErrDiscoveryUnsupported means the backend does not offer machine-login discovery:
	// an older backend without the route, or unexpected response content.
	ErrDiscoveryUnsupported = errors.New("backend does not support machine login; upgrade the backend (or use --token-endpoint and --audience)")

	// ErrDiscoveryUnavailable means the gateway was reached but its OpenID Connect
	// configuration is not currently available (HTTP 503); retry shortly.
	ErrDiscoveryUnavailable = errors.New("machine-login discovery is temporarily unavailable; retry shortly")

	// ErrNoTokenEndpoint means no token endpoint was provided via discovery or override.
	ErrNoTokenEndpoint = errors.New("no token endpoint available; specify one with --token-endpoint")

	// ErrNoAudience means no audience was provided from discovery or override.
	ErrNoAudience = errors.New("no audience available; specify one with --audience")

	// ErrMultipleAudiences means discovery listed several audiences and none was chosen.
	ErrMultipleAudiences = errors.New("discovery returned multiple audiences; choose one with --audience")
)

// fetchOidcDiscovery retrieves and validates the machine-login discovery document from the backend. It returns ErrDiscoveryUnsupported if the backend does not support machine login.
func fetchOidcDiscovery(ctx context.Context, storeUrl string) (*OidcDiscovery, error) {
	base := createHttpClient(storeUrl).BaseURL
	discUrl := base + "auth/oidc-discovery"

	ctx, cancel := context.WithTimeout(ctx, discoveryTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		// Route absent → older backend without machine-login support.
		return nil, ErrDiscoveryUnsupported
	case resp.StatusCode == http.StatusServiceUnavailable:
		return nil, ErrDiscoveryUnavailable
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		return nil, fmt.Errorf("machine-login discovery failed: HTTP %d", resp.StatusCode)
	}

	// Detect by response SHAPE, not just status:
	// a corporate proxy or a misconfigured ingress may answer an unknown route with 200 and an HTML page.
	// Require valid JSON that carries at least the token endpoint.
	var disc OidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&disc); err != nil {
		return nil, fmt.Errorf("machine-login discovery returned invalid JSON: %w", ErrDiscoveryUnsupported)
	}
	if strings.TrimSpace(disc.TokenEndpoint) == "" {
		return nil, fmt.Errorf("machine-login discovery returned JSON without tokenEndpoint: %w", ErrDiscoveryUnsupported)
	}
	return &disc, nil
}

// resolveTokenEndpoint returns the OIDC token endpoint to POST to, preferring an
// explicit --token-endpoint override over the discovered value.
func resolveTokenEndpoint(override string, disc *OidcDiscovery) (string, error) {
	if override != "" {
		return override, nil
	}
	if disc != nil && strings.TrimSpace(disc.TokenEndpoint) != "" {
		return disc.TokenEndpoint, nil
	}
	return "", ErrNoTokenEndpoint
}

// resolveAudience returns the API audience to request a machine token for,
// preferring an explicit --audience override.
// The backend does not yet expose a dedicated machine audience,
// so when discovery lists exactly one audience it is used; when it lists several, the caller must choose with --audience.
func resolveAudience(override string, disc *OidcDiscovery) (string, error) {
	if override != "" {
		return override, nil
	}
	if disc == nil || len(disc.Audiences) == 0 {
		return "", ErrNoAudience
	}
	if len(disc.Audiences) > 1 {
		return "", ErrMultipleAudiences
	}
	return disc.Audiences[0], nil
}
