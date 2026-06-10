package backend

import (
	"context"
	"m2cpcli/backend/legacy"
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
