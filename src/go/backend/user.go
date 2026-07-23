package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	v5 "m2cpcli/backend/v5"
	"m2cpcli/env"
	"m2cpcli/structs"

	"github.com/spf13/viper"
)

func GetUserById(ctx context.Context, id string) (*structs.User, error) {
	return legacy.GetUserById(ctx, id)
}

func GetUserByEmail(ctx context.Context, email string) (*structs.User, error) {
	return legacy.GetUserByEmail(ctx, email)
}

func GetCurrentUser(ctx context.Context) (*structs.User, error) {
	return legacy.GetCurrentUser(ctx)
}

func GetUserRolesByUserIdAndTenantId(ctx context.Context, userId string, tenantId string) (*legacy.GetUserRolesByUserIdAndTenantIdResponse, error) {
	return legacy.GetUserRolesByUserIdAndTenantId(ctx, userId, tenantId)
}

// GetMe has no legacy equivalent: it exists only to recover the tenantId for
// sessions authenticated via the external identity provider, which is a
// v5-only concern.
func GetMe(ctx context.Context) (*structs.User, error) {
	return v5.GetMe(ctx)
}

// GetMeWithFallback resolves the currently authenticated user, preferring the
// "me" query. One store's backend does not expose "me"; there, the identity is
// derived directly from the session JWT's claims, which on that tenant-scoped
// store carry tenant_id, role_values and scopes (see env.UserFromJwtClaims).
//
// The token fallback is only used when the JWT actually carries a tenant_id
// claim, so a genuine "me" failure on a store that does support it (network/
// auth errors, where the token has no such claim) surfaces the original error
// rather than being masked. Note the fallback yields neither a display
// name/isSuperAdmin flag (unknown without "me") - only what the token encodes.
func GetMeWithFallback(ctx context.Context) (*structs.User, error) {
	user, err := GetMe(ctx)
	if err == nil {
		return user, nil
	}

	fallbackUser, fallbackErr := env.UserFromJwtClaims(viper.GetString("jwt"))
	if fallbackErr != nil {
		// Fallback not applicable (e.g. external-provider token without a
		// tenant_id claim); surface the primary "me" error.
		return nil, err
	}
	return fallbackUser, nil
}
