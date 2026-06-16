package v5

import (
	"context"
	"fmt"
	"m2cpcli/structs"
	"m2cpcli/tools"
)

func GetSystemByChildAssetId(ctx context.Context, assetId string) (*structs.Asset, error) {
	res, err := getSystemByChildAssetId(ctx, assetId)
	if err != nil {
		return nil, err
	}
	if res != nil && res.Assets != nil {
		if len(res.Assets.Items) == 0 {
			return nil, nil
		}
		if len(res.Assets.Items) > 1 {
			return nil, fmt.Errorf("multiple matching Systems found")
		}
		resItem := res.Assets.Items[0]
		return GetSystemById(ctx, resItem.Id)
	}
	return nil, fmt.Errorf("unexpected empty result")
}

func GetSystemById(ctx context.Context, id string) (*structs.Asset, error) {
	res, err := getSystemInfoById(ctx, id)
	if err != nil {
		return nil, err
	}
	if res != nil && res.Assets != nil {
		if len(res.Assets.Items) == 0 {
			return nil, nil
		}
		if len(res.Assets.Items) > 1 {
			return nil, fmt.Errorf("multiple matching Systems found")
		}
		resItem := res.Assets.Items[0]
		system := structs.Asset{
			Id:                 resItem.Id,
			Serial:             resItem.SerialNo,
			McuId:              resItem.McuId,
			Components:         []structs.Asset{},
			IsSystemOwned:      &resItem.IsSystemOwned,
			Name:               resItem.AssetName,
			Description:        resItem.AssetDescription,
			SystemAssetId:      tools.Ptr(resItem.Id),
			ParentAssetId:      resItem.ParentAssetId,
			AssetType:          tools.Ptr(structs.AssetModelTypeIdToType(resItem.AssetModel.AssetType.Id)),
			AssetModelId:       resItem.AssetModel.Id,
			AssetModelName:     resItem.AssetModel.AssetModelName,
			AssetModelTypeId:   resItem.AssetModel.AssetType.Id,
			AssetModelTypeName: resItem.AssetModel.AssetType.AssetTypeName,
			TenantId:           resItem.Tenant.Id,
			TenantName:         resItem.Tenant.Alias,
		}
		for _, child := range resItem.ChildAssets {
			component := structs.Asset{
				SystemAssetId:          tools.StrPtr(resItem.Id),
				ParentAssetId:          tools.StrPtr(resItem.Id),
				AssetType:              tools.Ptr(structs.AssetModelTypeIdToType(child.AssetModel.AssetType.Id)),
				Id:                     child.Id,
				Serial:                 child.SerialNo,
				McuId:                  child.McuId,
				Name:                   child.AssetName,
				DeviceId:               child.DeviceId,
				AssetModelId:           child.AssetModel.Id,
				AssetModelName:         child.AssetModel.AssetModelName,
				AssetModelTypeCategory: formatAssetTypeCategory(child.AssetModel.AssetType.AssetTypeCategory),
				AssetModelTypeName:     child.AssetModel.AssetType.AssetTypeName,
				TenantId:               child.Tenant.Id,
				TenantName:             child.Tenant.Alias,
			}
			if child.DeviceId != nil || child.Device != nil || child.AttestationKey != nil {
				if component.Device == nil {
					component.Device = &structs.Device{}
				}
				if child.DeviceId != nil {
					component.Device.DeviceId = *child.DeviceId
				}
				if child.Device != nil {
					component.Device.DeviceSerial = child.Device.SerialNumber
				}
				if child.AttestationKey != nil {
					component.Device.DeviceAttestationKey = child.AttestationKey
				}
			}
			system.Components = append(system.Components, component)

		}

		return &system, nil
	}
	return nil, fmt.Errorf("unexpected empty result")
}

func formatAssetTypeCategory(assetType AssetTypeCategory) string {
	if assetType == "EDGE_DEVICE" {
		return "Edge Device"
	} else if assetType == "REAL_TIME_DEVICE" {
		return "Real Time Device"
	} else {
		return "invalid"
	}
}
