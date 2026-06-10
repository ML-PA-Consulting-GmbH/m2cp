package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	v5 "m2cpcli/backend/v5"
	"m2cpcli/structs"
)

func GetSystemByChildAssetId(ctx context.Context, assetId string) (*structs.Asset, error) {
	if backendMajorVersionInt(ctx) < 5 {
		return legacy.GetSystemByChildAssetId(ctx, assetId)
	}
	return v5.GetSystemByChildAssetId(ctx, assetId)
}

func GetSystemById(ctx context.Context, id string) (*structs.Asset, error) {
	if backendMajorVersionInt(ctx) < 5 {
		return legacy.GetSystemById(ctx, id)
	}
	return v5.GetSystemById(ctx, id)
}
