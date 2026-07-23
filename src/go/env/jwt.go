package env

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"m2cpcli/structs"
	"strings"
	"time"
)

type JsonWebToken struct {
	ExpirationTime float64
}

// DecodeJwtClaims decodes the payload segment of a JWT into a claims map.
func DecodeJwtClaims(jwt string) (map[string]interface{}, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("could not split JWT")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode JWT")
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("could not unmarshall JWT")
	}
	return claims, nil
}

func NewJsonWebToken(jwt string) (*JsonWebToken, error) {
	if jwt == "" {
		return nil, nil
	}

	claims, err := DecodeJwtClaims(jwt)
	if err != nil {
		return nil, err
	}

	result := JsonWebToken{}

	var ok bool
	result.ExpirationTime, ok = claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("could not find ExpirationTime")
	}
	return &result, nil
}

// UserFromJwtClaims derives a user identity directly from a JWT's claims.
//
// It is the fallback identity source for stores whose backend does not expose
// the "me" query (which is otherwise preferred, see backend.GetMeWithFallback):
// on such tenant-scoped stores the token itself carries the "tenant_id",
// "role_values" and "scopes" claims. Each of role_values/scopes may be encoded
// as a single string or an array of strings.
//
// An error is returned when the token carries no "tenant_id" claim. This is the
// discriminator that makes the fallback safe: tokens issued by an external
// identity provider (browser-based login) lack that claim, so the caller keeps
// the original "me" error instead of masking it. IsSuperAdmin is intentionally
// left nil - no JWT carries that claim; it is only known once "me" has run.
func UserFromJwtClaims(jwt string) (*structs.User, error) {
	claims, err := DecodeJwtClaims(jwt)
	if err != nil {
		return nil, err
	}

	tenantId, _ := claims["tenant_id"].(string)
	if tenantId == "" {
		return nil, fmt.Errorf("token carries no tenant_id claim")
	}

	user := &structs.User{
		TenantId: tenantId,
		Permissions: &structs.Permissions{
			Roles:  claimStringSlice(claims["role_values"]),
			Scopes: claimStringSlice(claims["scopes"]),
		},
	}
	// Some providers also embed the user's email in the token; use it if present.
	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}
	return user, nil
}

// claimStringSlice normalizes a JWT claim that may be either a single string or
// an array of strings into a []string (nil for any other/absent shape).
func claimStringSlice(claim interface{}) []string {
	switch v := claim.(type) {
	case string:
		return []string{v}
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

func (jwt *JsonWebToken) IsValid() bool {
	if jwt == nil {
		return false
	}
	expiresTime := time.Unix(int64(jwt.ExpirationTime), 0)
	if expiresTime.Before(time.Now()) {
		return false
	}
	return true
}
