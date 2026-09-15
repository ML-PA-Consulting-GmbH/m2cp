package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// clientSecretEnvVar is the environment variable the CLI reads the M2M client secret from
// when --client-secret-stdin is not used.
const clientSecretEnvVar = "M2CP_CLIENT_SECRET"

// ValidateM2MLoginParams checks the required parameters for a client-credentials ("--method m2m") login.
// An org id is required because every machine token must be organization-scoped.
func ValidateM2MLoginParams(clientId, orgId string) error {
	if clientId == "" {
		return fmt.Errorf("--client-id is required for --method m2m")
	}
	if orgId == "" {
		return fmt.Errorf("--org-id is required for --method m2m")
	}
	return nil
}

// ResolveClientSecret returns the client secret for a client-credentials login.
// Resolved by order of priority:
//  1. stdin (if fromStdin is true)
//  2. M2CP_CLIENT_SECRET environment variable
func ResolveClientSecret(fromStdin bool, stdin io.Reader) (string, error) {
	if fromStdin {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read client secret from stdin: %w", err)
		}
		// Strip only a trailing newline (as produced by `echo`/pipes); keep any
		// other characters the secret may legitimately contain.
		secret := strings.TrimRight(string(data), "\r\n")
		if secret == "" {
			return "", fmt.Errorf("no client secret provided on stdin")
		}
		return secret, nil
	}

	secret := os.Getenv(clientSecretEnvVar)
	if secret == "" {
		return "", fmt.Errorf("client secret required: set %s or pass --client-secret-stdin", clientSecretEnvVar)
	}
	return secret, nil
}

const tokenRequestTimeout = 30 * time.Second

// M2MLoginParams holds the inputs for a client-credentials token request.
type M2MLoginParams struct {
	TokenEndpoint string
	ClientID      string
	ClientSecret  string
	Audience      string
	OrgID         string
}

// m2mTokenResponse is the OAuth2 token-endpoint success response.
type m2mTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// acquireM2MToken performs the OAuth2 client-credentials grant against the
// token endpoint and returns the access token. The organization parameter is
// always sent: machine tokens are organization-scoped in this design.
//
// The client secret is only ever sent in the request body; it is never logged
// or included in a returned error.
func acquireM2MToken(ctx context.Context, p M2MLoginParams) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", p.ClientID)
	form.Set("client_secret", p.ClientSecret)
	form.Set("audience", p.Audience)  // required for auth0 with no default audience configured
	form.Set("organization", p.OrgID) // required for (our expected) auth0 setup

	ctx, cancel := context.WithTimeout(ctx, tokenRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// OAuth error responses carry {error, error_description}; surface those
		// (never the request itself, which holds the secret).
		return "", fmt.Errorf("token request rejected (HTTP %d): %s", resp.StatusCode, tokenErrorMessage(body))
	}

	var tr m2mTokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}
	if tr.AccessToken == "" {
		return "", errors.New("token response did not include an access_token")
	}
	return tr.AccessToken, nil
}

// tokenErrorMessage extracts a helpful message from an OAuth2 error-response body
// ({error, error_description}), falling back to a trimmed snippet of the body.
func tokenErrorMessage(body []byte) string {
	var e struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if json.Unmarshal(body, &e) == nil && e.Error != "" {
		if e.Description != "" {
			return fmt.Sprintf("%s: %s", e.Error, e.Description)
		}
		return e.Error
	}
	snippet := strings.TrimSpace(string(body))
	if len(snippet) > 200 {
		snippet = snippet[:200]
	}
	return snippet
}

// M2MLoginInput holds everything needed to obtain a machine token: the store
// URL (for discovery), the client credentials, and optional discovery overrides.
type M2MLoginInput struct {
	StoreURL              string
	ClientID              string
	OrgID                 string
	Secret                string
	TokenEndpointOverride string
	AudienceOverride      string
}

// M2MLogin performs a full machine (client-credentials) login:
// it discovers the token endpoint and audience (skipped when both are overridden),
// then acquires an organization-scoped access token.
func M2MLogin(ctx context.Context, in M2MLoginInput) (string, error) {
	var disc *OidcDiscovery
	if in.TokenEndpointOverride == "" || in.AudienceOverride == "" {
		var err error
		if disc, err = fetchOidcDiscovery(ctx, in.StoreURL); err != nil {
			return "", err
		}
	}

	tokenEndpoint, err := resolveTokenEndpoint(in.TokenEndpointOverride, disc)
	if err != nil {
		return "", err
	}
	audience, err := resolveAudience(in.AudienceOverride, disc)
	if err != nil {
		return "", err
	}

	return acquireM2MToken(ctx, M2MLoginParams{
		TokenEndpoint: tokenEndpoint,
		ClientID:      in.ClientID,
		ClientSecret:  in.Secret,
		Audience:      audience,
		OrgID:         in.OrgID,
	})
}
