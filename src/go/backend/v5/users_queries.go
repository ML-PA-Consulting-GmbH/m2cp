package v5

import (
	"context"
	"fmt"
	"m2cpcli/structs"
)

func GetUserById(ctx context.Context, id string) (*structs.User, error) {
	res, err := getUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	if res.User == nil {
		return nil, fmt.Errorf("User %s not found", id)
	}
	return &structs.User{
		Id:         res.User.Id,
		Name:       res.User.DisplayName,
		Email:      res.User.Email,
		SshKey:     res.User.SshPublicKey,
		TenantId:   res.User.TenantId,
		TenantName: res.User.Tenant.TenantName,
	}, nil
}

func GetUserByEmail(ctx context.Context, email string) (*structs.User, error) {
	res, err := getUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if res.Users == nil || res.Users.Items == nil || len(res.Users.Items) != 1 {
		return nil, fmt.Errorf("User %s not found", email)
	}
	return &structs.User{
		Id:         res.Users.Items[0].Id,
		Name:       res.Users.Items[0].DisplayName,
		Email:      res.Users.Items[0].Email,
		SshKey:     res.Users.Items[0].SshPublicKey,
		TenantId:   res.Users.Items[0].TenantId,
		TenantName: res.Users.Items[0].Tenant.TenantName,
	}, nil
}

func GetCurrentUser(ctx context.Context) (*structs.User, error) {
	res, err := getCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return &structs.User{
		Id:         res.CurrentUser.Id,
		Name:       res.CurrentUser.DisplayName,
		Email:      res.CurrentUser.Email,
		SshKey:     res.CurrentUser.SshPublicKey,
		TenantId:   res.CurrentUser.TenantId,
		TenantName: res.CurrentUser.Tenant.TenantName,
	}, nil
}

// GetMe fetches the currently authenticated user, including the tenantId.
// This is required after browser-based login, since the JWT issued by the
// external authentication provider does not carry a tenant_id claim.
func GetMe(ctx context.Context) (*structs.User, error) {
	res, err := me(ctx)
	if err != nil {
		return nil, err
	}
	if res.Me == nil {
		return nil, fmt.Errorf("me query returned no user")
	}

	user := &structs.User{
		Id:       res.Me.Id,
		Name:     res.Me.DisplayName,
		Email:    res.Me.Email,
		TenantId: res.Me.TenantId,
	}
	if res.Me.Permissions != nil {
		user.Permissions = &structs.Permissions{
			Scopes:       res.Me.Permissions.Scopes,
			Roles:        res.Me.Permissions.Roles,
			IsSuperAdmin: &res.Me.Permissions.IsSuperAdmin,
		}
	}
	return user, nil
}
