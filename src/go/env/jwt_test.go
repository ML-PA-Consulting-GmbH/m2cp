package env

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

// makeJwt builds a syntactically valid JWT (header.payload.signature) whose
// payload encodes the given claims. Signature is irrelevant here: the code
// under test only base64-decodes and JSON-parses the payload segment.
func makeJwt(t *testing.T, claims map[string]interface{}) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	assert.NoError(t, err)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}

func TestJsonWebToken_IsValid(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	str := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJkaXNwbGF5X25hbWUiOiJKYWtvYiBEdWViZWwiLCJlbWFpbCI6Impha29iLmR1ZWJlbEBtbC1wYS5jb20iLCJzY29wZXMiOiIqLiouQWxsIiwic2Vzc2lvbl9pZCI6IjY0ZmVhZTg1LWZiY2ItNDE0NS1iYWMzLTE0Mzg3ZWIzNDdkYSIsInRlbmFudF9pZCI6IjIyNmRmODEwLWIxNDYtNDU5ZS1iYzY0LWRlZmRmYzVlNjA2MiIsInVzZXJfaWQiOiJlZmEzOGNmMC1jOTJlLTQzMzktYWNjMC1hY2YxMDE1YWI5YWEiLCJuYmYiOjE2OTM4NDMyNjgsImV4cCI6MTY5Mzg4NjQ2OCwiaWF0IjoxNjkzODQzMjY4LCJpc3MiOiJodHRwczovL3d3dy5tbC1wYS5jb20vIn0.fREXUAR7N-RnHBQARy56pGGfVGpQEgADFR0J2nVk46Ux-99s4kotNXU4WqnE8fx-6YhGlOoBRG0fmtmd5ObS4lH1LiGi5PT4Qsh4j6zdELp3rPkvJvBhMCd4JUMXmW3BEES1QM_-HCKtmgF2O58MdNHjq3vXLfl9rGSuTeFP-QJJtfd54kx6MDZdyC8LXz5iIPXclA7FACDgZwD2T9IV5EUoHK9cxGSdwo-6cVgQQg_Xnw_l0qKceNlzK9fXGKtcK2qA-9zeNuxPOOS01KgwRlYgLKxgdqCStcG1e_6OSRln2nRUgWpXrqc2ob5J1blRKe2wnJH7VcLKhpdLKu98zg`
	jwt, err := NewJsonWebToken(str)
	assert.NoError(t, err)
	assert.NotNil(t, jwt)

	assert.False(t, jwt.IsValid())
}

func TestDecodeJwtClaims_MalformedToken(t *testing.T) {
	defer goleak.VerifyNone(t)

	claims, err := DecodeJwtClaims("not-a-jwt")
	assert.Error(t, err)
	assert.Nil(t, claims)

	claims, err = DecodeJwtClaims("")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestUserFromJwtClaims_ScopeAsSingleString(t *testing.T) {
	defer goleak.VerifyNone(t)

	// This is the real token used above: tenant_id set, scopes as a single
	// string ("*.*.All"), email present, and no role_values claim.
	str := `eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJkaXNwbGF5X25hbWUiOiJKYWtvYiBEdWViZWwiLCJlbWFpbCI6Impha29iLmR1ZWJlbEBtbC1wYS5jb20iLCJzY29wZXMiOiIqLiouQWxsIiwic2Vzc2lvbl9pZCI6IjY0ZmVhZTg1LWZiY2ItNDE0NS1iYWMzLTE0Mzg3ZWIzNDdkYSIsInRlbmFudF9pZCI6IjIyNmRmODEwLWIxNDYtNDU5ZS1iYzY0LWRlZmRmYzVlNjA2MiIsInVzZXJfaWQiOiJlZmEzOGNmMC1jOTJlLTQzMzktYWNjMC1hY2YxMDE1YWI5YWEiLCJuYmYiOjE2OTM4NDMyNjgsImV4cCI6MTY5Mzg4NjQ2OCwiaWF0IjoxNjkzODQzMjY4LCJpc3MiOiJodHRwczovL3d3dy5tbC1wYS5jb20vIn0.fREXUAR7N-RnHBQARy56pGGfVGpQEgADFR0J2nVk46Ux-99s4kotNXU4WqnE8fx-6YhGlOoBRG0fmtmd5ObS4lH1LiGi5PT4Qsh4j6zdELp3rPkvJvBhMCd4JUMXmW3BEES1QM_-HCKtmgF2O58MdNHjq3vXLfl9rGSuTeFP-QJJtfd54kx6MDZdyC8LXz5iIPXclA7FACDgZwD2T9IV5EUoHK9cxGSdwo-6cVgQQg_Xnw_l0qKceNlzK9fXGKtcK2qA-9zeNuxPOOS01KgwRlYgLKxgdqCStcG1e_6OSRln2nRUgWpXrqc2ob5J1blRKe2wnJH7VcLKhpdLKu98zg`

	user, err := UserFromJwtClaims(str)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "226df810-b146-459e-bc64-defdfc5e6062", user.TenantId)
	assert.Equal(t, "jakob.duebel@ml-pa.com", user.Email)
	assert.NotNil(t, user.Permissions)
	assert.Equal(t, []string{"*.*.All"}, user.Permissions.Scopes)
	assert.Empty(t, user.Permissions.Roles)
	assert.Nil(t, user.Permissions.IsSuperAdmin, "isSuperAdmin is never derivable from a token")
}

func TestUserFromJwtClaims_RolesAndScopesAsArrays(t *testing.T) {
	defer goleak.VerifyNone(t)

	str := makeJwt(t, map[string]interface{}{
		"tenant_id":   "tenant-123",
		"role_values": []string{"admin", "operator"},
		"scopes":      []string{"device.read", "device.write"},
	})

	user, err := UserFromJwtClaims(str)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "tenant-123", user.TenantId)
	assert.Equal(t, []string{"admin", "operator"}, user.Permissions.Roles)
	assert.Equal(t, []string{"device.read", "device.write"}, user.Permissions.Scopes)
}

func TestUserFromJwtClaims_NoTenantIdRejected(t *testing.T) {
	defer goleak.VerifyNone(t)

	// A token from an external identity provider carries no tenant_id claim;
	// the fallback must reject it so the caller keeps the primary "me" error.
	str := makeJwt(t, map[string]interface{}{
		"email": "someone@example.com",
		"exp":   1693886468,
	})

	user, err := UserFromJwtClaims(str)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestJsonWebToken_EmptyStringIsValid(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	str := ""
	jwt, err := NewJsonWebToken(str)
	assert.NoError(t, err)
	assert.Nil(t, jwt)
	assert.False(t, jwt.IsValid())
}
