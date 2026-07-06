package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	v5 "m2cpcli/backend/v5"
	"m2cpcli/structs"
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
