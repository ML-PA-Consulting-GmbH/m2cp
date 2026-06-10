package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	v5 "m2cpcli/backend/v5"
	"m2cpcli/structs"
)

func SetAssetSystem(ctx context.Context, deviceAssetId string, systemAssetId *string) error {
	return legacy.SetAssetSystem(ctx, deviceAssetId, systemAssetId)
}

func CreateAsset(ctx context.Context, modelId string, hwSerial string, mcuId *string, assetName string, assetDescription *string, attestationKey *string) (id string, err error) {
	if backendMajorVersionInt(ctx) < 5 {
		return legacy.CreateAsset(ctx, modelId, hwSerial, assetName, assetDescription, attestationKey)
	}
	return v5.CreateAsset(ctx, modelId, hwSerial, mcuId, assetName, assetDescription, attestationKey)
}

func CreateAssetModel(ctx context.Context, modelTypeId string, name string, description *string) (id string, err error) {
	return legacy.CreateAssetModel(ctx, modelTypeId, name, description)
}

func GetAssetModels(ctx context.Context, types []string) ([]structs.AssetModel, error) {
	return legacy.GetAssetModels(ctx, types)
}

func GetAssetModelById(ctx context.Context, assetModelId string) (*structs.AssetModel, error) {
	return legacy.GetAssetModelById(ctx, assetModelId)
}

func FindAsset(ctx context.Context, q string) ([]structs.Asset, error) {
	if backendMajorVersionInt(ctx) < 5 {
		return legacy.FindAsset(ctx, q)
	}
	return v5.FindAsset(ctx, q)
}

func GetAssetById(ctx context.Context, q string) (*structs.Asset, error) {
	if backendMajorVersionInt(ctx) < 5 {
		return legacy.GetAssetById(ctx, q)
	}
	return v5.GetAssetById(ctx, q)
}

func GetSystemsList(ctx context.Context, filter []structs.BackendQueryFilter) (assets []structs.Asset, err error) {
	if backendMajorVersionInt(ctx) < 5 {
		return legacy.GetSystemsList(ctx, filter)
	}
	return v5.GetSystemsList(ctx, filter)
}

func UpdateAssetModel(ctx context.Context, id string, name *string, description *string) (err error) {
	return legacy.UpdateAssetModel(ctx, id, name, description)
}

func UpdateAsset(ctx context.Context, id string, name *string, attestationKey *string) (err error) {
	return legacy.UpdateAsset(ctx, id, name, attestationKey)
}

func ProvisionSystemAsset(ctx context.Context, input structs.SystemAssetProvisionInput) (*structs.SystemAssetProvisionOutput, error) {
	if backendMajorVersionInt(ctx) < 5 {
		return legacy.ProvisionSystemAsset(ctx, &input)
	}
	return v5.ProvisionSystemAsset(ctx, &input)
}
