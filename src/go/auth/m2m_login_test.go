package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateM2MLoginParams(t *testing.T) {
	assert.Error(t, ValidateM2MLoginParams("", "org_1"), "missing client id must error")
	assert.Error(t, ValidateM2MLoginParams("client_1", ""), "missing org id must error")
	assert.NoError(t, ValidateM2MLoginParams("client_1", "org_1"))
}

func TestResolveClientSecret_StdinWinsAndTrimsNewline(t *testing.T) {
	t.Setenv("M2CP_CLIENT_SECRET", "from-env")
	secret, err := ResolveClientSecret(true, strings.NewReader("from-stdin\n"))
	assert.NoError(t, err)
	assert.Equal(t, "from-stdin", secret, "stdin takes precedence and the trailing newline is stripped")
}

func TestResolveClientSecret_StdinEmpty(t *testing.T) {
	_, err := ResolveClientSecret(true, strings.NewReader("\n"))
	assert.Error(t, err)
}

func TestResolveClientSecret_FromEnv(t *testing.T) {
	t.Setenv("M2CP_CLIENT_SECRET", "s3cr3t")
	secret, err := ResolveClientSecret(false, strings.NewReader(""))
	assert.NoError(t, err)
	assert.Equal(t, "s3cr3t", secret)
}

func TestResolveClientSecret_MissingEnv(t *testing.T) {
	t.Setenv("M2CP_CLIENT_SECRET", "")
	_, err := ResolveClientSecret(false, strings.NewReader(""))
	assert.Error(t, err)
}

func TestAcquireM2MToken_Success(t *testing.T) {
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.Form
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"header.payload.sig","token_type":"Bearer","expires_in":3600}`))
	}))
	defer srv.Close()

	token, err := acquireM2MToken(t.Context(), M2MLoginParams{
		TokenEndpoint: srv.URL,
		ClientID:      "client_1",
		ClientSecret:  "topsecret-DO-NOT-LEAK",
		Audience:      testAudience,
		OrgID:         "org_1",
	})
	require.NoError(t, err)
	assert.Equal(t, "header.payload.sig", token)
	// the exact client-credentials form, including the mandatory organization
	assert.Equal(t, "client_credentials", gotForm.Get("grant_type"))
	assert.Equal(t, "client_1", gotForm.Get("client_id"))
	assert.Equal(t, "topsecret-DO-NOT-LEAK", gotForm.Get("client_secret"))
	assert.Equal(t, testAudience, gotForm.Get("audience"))
	assert.Equal(t, "org_1", gotForm.Get("organization"))
}

func TestAcquireM2MToken_OAuthErrorIsSurfacedWithoutSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"Client authentication failed"}`))
	}))
	defer srv.Close()

	_, err := acquireM2MToken(t.Context(), M2MLoginParams{
		TokenEndpoint: srv.URL,
		ClientID:      "client_1",
		ClientSecret:  "topsecret-DO-NOT-LEAK",
		Audience:      testAudience,
		OrgID:         "org_1",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid_client")
	assert.ErrorContains(t, err, "Client authentication failed")
	assert.NotContains(t, err.Error(), "topsecret-DO-NOT-LEAK", "the client secret must never appear in an error")
}

func TestAcquireM2MToken_MissingAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	_, err := acquireM2MToken(t.Context(), M2MLoginParams{
		TokenEndpoint: srv.URL,
		ClientID:      "c",
		ClientSecret:  "x",
		Audience:      testAudience,
		OrgID:         "o",
	})
	assert.ErrorContains(t, err, "access_token")
}

func TestM2MLogin_ViaDiscovery(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/auth/oidc-discovery", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// token endpoint points back at this same test server
		_, _ = w.Write([]byte(`{
			"providerName": "AUTH0",
			"issuers": ["` + testIssuer + `"],
			"audiences": ["` + testAudience + `"],
			"tokenEndpoint": "` + srv.URL + `/oauth/token"
		}`))
	})
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"header.payload.sig","token_type":"Bearer","expires_in":3600}`))
	})

	token, err := M2MLogin(t.Context(), M2MLoginInput{
		StoreURL: srv.URL + "/graphql",
		ClientID: "client_1",
		OrgID:    "org_1",
		Secret:   "s3cr3t",
	})
	require.NoError(t, err)
	assert.Equal(t, "header.payload.sig", token)
}

func TestM2MLogin_WithOverrides_SkipsDiscovery(t *testing.T) {
	tokenHit := false
	mux := http.NewServeMux()
	// The discovery route is intentionally absent (ServeMux answers 404); if it
	// were called, M2MLogin would fail — proving the overrides skip discovery.
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		tokenHit = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"header.payload.sig","token_type":"Bearer"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	token, err := M2MLogin(t.Context(), M2MLoginInput{
		StoreURL:              srv.URL + "/graphql",
		ClientID:              "client_1",
		OrgID:                 "org_1",
		Secret:                "s3cr3t",
		TokenEndpointOverride: srv.URL + "/oauth/token",
		AudienceOverride:      testAudience,
	})
	require.NoError(t, err)
	assert.Equal(t, "header.payload.sig", token)
	assert.True(t, tokenHit, "token endpoint should be called")
}
