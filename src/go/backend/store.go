package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	"m2cpcli/backend/v5"
	"m2cpcli/structs"

	"github.com/Masterminds/semver/v3"
)

var backendMajorVersionCache *int

var backendStoreVersionCache *string

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

// backendAtLeast reports whether the backend's store version is >= minVersion
// (a semver string). It fails closed: if the version cannot be resolved (e.g. a
// legacy backend without the v5 store-info query) or parsed, it returns false,
// so a version-gated feature is never sent to a backend that may not support it.
func backendAtLeast(ctx context.Context, minVersion string) bool {
	if backendStoreVersionCache == nil {
		v, err := v5.BackendStoreVersion(ctx)
		if err != nil {
			return false
		}
		backendStoreVersionCache = &v
	}
	current, err := semver.NewVersion(*backendStoreVersionCache)
	if err != nil {
		return false
	}
	minimum, err := semver.NewVersion(minVersion)
	if err != nil {
		return false
	}
	return !current.LessThan(minimum)
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
