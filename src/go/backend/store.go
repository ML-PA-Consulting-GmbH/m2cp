package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	"m2cpcli/backend/v5"
	"m2cpcli/structs"
)

var backendMajorVersionCache *int

// backendMajorVersionInt is a tool function for switching between backend versions
func backendMajorVersionInt(ctx context.Context) int {
	if backendMajorVersionCache == nil {
		x, err := BackendMajorVersion(ctx)
		if err != nil {
			return -1
		}
		backendMajorVersionCache = &x
	}
	return *backendMajorVersionCache
}

func GetStoreOwner(ctx context.Context) (tenantName, tenantId string, err error) {
	return legacy.GetStoreOwner(ctx)
}

func GetAssetsOwner(ctx context.Context) (tenantName, tenantId string, err error) {
	return legacy.GetAssetsOwner(ctx)
}

func BackendInfo(ctx context.Context) (*structs.Backend, error) {
	res, err := v5.BackendInfo(ctx)
	if err == nil {
		return res, nil
	}
	return legacy.BackendInfo(ctx)
}

func BackendMajorVersion(ctx context.Context) (int, error) {
	return v5.BackendMajorVersion(ctx)
}
