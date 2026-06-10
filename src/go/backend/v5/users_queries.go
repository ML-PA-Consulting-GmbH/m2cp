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
