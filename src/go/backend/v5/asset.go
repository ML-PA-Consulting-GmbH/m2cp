package v5

import (
	"context"
	"errors"
	"fmt"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"strings"
)

func CreateAsset(ctx context.Context, modelId string, hwSerial string, mcuId *string, assetName string, assetDescription *string, attestationKey *string) (id string, err error) {
	_, ownerId, err := GetAssetsOwner(ctx)
	if err != nil {
		return "", err
	}
	res, err := createAsset(ctx, modelId, hwSerial, mcuId, assetName, assetDescription, attestationKey, ownerId)
	if err != nil {
		return "", err
	}
	if res.CreateAssets == nil || len(res.CreateAssets) == 0 {
		return "", fmt.Errorf("query failed with empty result")
	}
	return res.CreateAssets[0].Id, nil
}

func FindAsset(ctx context.Context, q string) ([]structs.Asset, error) {
	q = strings.ToLower(q)
	q = strings.TrimSpace(q)
	q = strings.TrimPrefix(q, "sn")
	q = strings.TrimPrefix(q, "#")
	q = strings.TrimPrefix(q, "sn")
	q = strings.TrimLeft(q, " ")

	res, err := findAsset(ctx, q)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("query failed with nil result")
	}
	if res.Assets == nil || res.Assets.Items == nil {
		return nil, errors.New("query failed with nil result")
	}
	assets := make([]structs.Asset, len(res.Assets.Items))
	for i, asset := range res.Assets.Items {
		component := structs.Asset{
			Id:                     asset.Id,
			Name:                   tools.MaybeStringToString(asset.AssetName, ""),
			Serial:                 asset.SerialNo,
			McuId:                  asset.McuId,
			AssetModelId:           asset.AssetModel.Id,
			AssetModelName:         asset.AssetModel.AssetModelName,
			AssetModelTypeId:       asset.AssetModel.AssetType.Id,
			AssetModelTypeName:     asset.AssetModel.AssetType.AssetTypeName,
			AssetModelTypeCategory: string(asset.AssetModel.AssetType.AssetTypeCategory),
			DeviceId:               asset.DeviceId,
			TenantId:               asset.Tenant.Id,
			TenantName:             asset.Tenant.Alias,
		}
		component.AssetType = tools.Ptr(structs.AssetModelTypeIdToType(asset.AssetModel.AssetType.Id))
		if asset.IsSystemOwned && asset.ParentAssetId != nil {
			component.SystemAssetId = asset.ParentAssetId
			component.ParentAssetId = asset.ParentAssetId
		} else if component.AssetModelTypeName == "System" {
			component.SystemAssetId = &component.Id
		}
		if asset.DeviceId != nil {
			component.Device = &structs.Device{
				DeviceId:     *asset.DeviceId,
				DeviceSerial: asset.Device.SerialNumber,
			}
		}
		assets[i] = component
	}

	if tools.IsValidUuid(q) {
		asset, err := GetAssetById(ctx, q)
		if err == nil && asset != nil {
			assets = append(assets, *asset)
		}

	}
	return assets, nil
}

func GetAssetById(ctx context.Context, q string) (*structs.Asset, error) {
	res, err := getAssetById(ctx, q)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("no Assets found")
	}
	if res.Assets == nil || res.Assets.Items == nil || len(res.Assets.Items) == 0 {
		return nil, errors.New("no Assets found")
	}
	asset := res.Assets.Items[0]
	component := structs.Asset{
		Id:                     asset.Id,
		Name:                   tools.MaybeStringToString(asset.AssetName, ""),
		Serial:                 asset.SerialNo,
		McuId:                  asset.McuId,
		AssetModelName:         asset.AssetModel.AssetModelName,
		AssetModelId:           asset.AssetModel.Id,
		AssetType:              tools.StrPtr(structs.AssetModelTypeIdToType(asset.AssetModel.AssetTypeId)),
		AssetModelTypeName:     asset.AssetModel.AssetType.AssetTypeName,
		AssetModelTypeCategory: string(asset.AssetModel.AssetType.AssetTypeCategory),
		TenantId:               asset.Tenant.Id,
		TenantName:             asset.Tenant.Alias,
		DeviceId:               asset.DeviceId,
	}
	component.AssetType = tools.Ptr(structs.AssetModelTypeIdToType(asset.AssetModel.AssetType.Id))
	if asset.IsSystemOwned && asset.ParentAssetId != nil {
		component.SystemAssetId = asset.ParentAssetId
		component.ParentAssetId = asset.ParentAssetId
	} else if component.AssetModelTypeName == "System" {
		component.SystemAssetId = &component.Id
	}
	if asset.DeviceId != nil {
		component.Device = &structs.Device{
			DeviceSerial:         asset.Device.SerialNumber,
			DeviceAttestationKey: asset.AttestationKey,
		}
	}
	return &component, nil
}

func patternToFilterInput(pattern string) *StringOperationFilterInput {
	trimmed := strings.Trim(pattern, "*")
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return &StringOperationFilterInput{Contains: &trimmed}
	} else if strings.HasPrefix(pattern, "*") {
		return &StringOperationFilterInput{EndsWith: &trimmed}
	} else if strings.HasSuffix(pattern, "*") {
		return &StringOperationFilterInput{StartsWith: &trimmed}
	}
	return &StringOperationFilterInput{Eq: &trimmed}
}

func makeAssetFilterCondition(field string, filterInput *StringOperationFilterInput) *AssetFilterInput {
	switch field {
	case "name":
		return &AssetFilterInput{AssetName: filterInput}
	}
	panic("unknown field: " + field)
}

func backendQueryFiltersToAssetFilterAndSort(filters []structs.BackendQueryFilter) (filter *AssetFilterInput, sort []*AssetSortInput, err error) {
	conditions := make([]*AssetFilterInput, 0)
	assetSort := make([]*AssetSortInput, 0)
	for _, f := range filters {
		if f.Pattern != nil {
			conditions = append(conditions, makeAssetFilterCondition(f.Field, patternToFilterInput(*f.Pattern)))
		}
		if len(f.ExactAny) > 0 {
			typesFilter := make([]*AssetFilterInput, 0)
			for _, t := range f.ExactAny {
				typesFilter = append(typesFilter, &AssetFilterInput{AssetModel: &AssetModelFilterInput{AssetTypeId: &ComparableGuidOperationFilterInput{Eq: &t}}})
			}
			conditions = append(conditions, &AssetFilterInput{Or: typesFilter})
		}
		if f.Sort != nil {
			sortDirection := SortEnumType(*f.Sort)
			switch f.Field {
			case structs.BackendQueryFilterAssetId:
				assetSort = append(assetSort, &AssetSortInput{Id: &sortDirection})
			case structs.BackendQueryFilterAssetName:
				assetSort = append(assetSort, &AssetSortInput{AssetName: &sortDirection})
			case structs.BackendQueryFilterCreatedAt:
				assetSort = append(assetSort, &AssetSortInput{CreatedAt: &sortDirection})
			case structs.BackendQueryFilterModifiedAt:
				assetSort = append(assetSort, &AssetSortInput{ModifiedAt: &sortDirection})
			default:
				return nil, nil, fmt.Errorf("invalid sort key '%s'", f.Field)
			}
		}
	}

	return &AssetFilterInput{
		And: conditions,
	}, assetSort, nil
}

func GetSystemsList(ctx context.Context, filters []structs.BackendQueryFilter) (assets []structs.Asset, err error) {
	filter, sort, err := backendQueryFiltersToAssetFilterAndSort(filters)
	if err != nil {
		return nil, err
	}

	take := 100
	skip := 0
	hasNextPage := true
	for hasNextPage {
		var result *getAssetsResponse
		result, err = getAssets(ctx, filter, sort, &take, &skip)
		if err != nil {
			return nil, err
		}

		if result.Assets != nil && result.Assets.Items != nil && len(result.Assets.Items) > 0 {
			for _, resItem := range result.Assets.Items {
				if resItem.AssetModel.AssetType.AssetTypeName != "M2CP" && resItem.AssetModel.AssetType.AssetTypeName != "RTOS" && resItem.AssetModel.AssetType.AssetTypeName != "System" {
					continue
				}
				var asset = structs.Asset{
					Id:                     resItem.Id,
					Serial:                 resItem.SerialNo,
					McuId:                  resItem.McuId,
					Components:             []structs.Asset{},
					IsSystemOwned:          tools.Ptr(resItem.IsSystemOwned),
					Name:                   tools.MaybeStringToString(resItem.AssetName, ""),
					Description:            resItem.AssetDescription,
					ParentAssetId:          resItem.ParentAssetId,
					AssetModelId:           resItem.AssetModel.Id,
					AssetModelName:         resItem.AssetModel.AssetModelName,
					AssetModelTypeId:       resItem.AssetModel.AssetType.Id,
					AssetModelTypeName:     resItem.AssetModel.AssetType.AssetTypeName,
					AssetModelTypeCategory: string(resItem.AssetModel.AssetType.AssetTypeCategory),
					TenantId:               resItem.Tenant.Id,
					TenantName:             resItem.Tenant.Alias,
					DeviceId:               resItem.DeviceId,
				}
				asset.AssetType = tools.Ptr(structs.AssetModelTypeIdToType(resItem.AssetModel.AssetType.Id))
				assets = append(assets, asset)
			}
		}

		hasNextPage = result.Assets.PageInfo.HasNextPage
		skip += take
	}

	return assets, nil
}

func UpdateAssetModel(ctx context.Context, id string, name *string, description *string) (err error) {
	_, err = modifyAssetModel(ctx, id, name, description)
	return err
}

func UpdateAsset(ctx context.Context, id string, name *string, attestationKey *string) (err error) {
	_, err = modifyAsset(ctx, id, name, attestationKey)
	return err
}

func GetAssetsOwner(ctx context.Context) (tenantName, tenantId string, err error) {
	// TODO: this is another heuristic to solve the problem: under which tenant should we create assets for Systems
	res, err := BackendInfo(ctx)
	if err != nil {
		return "", "", err
	}
	if res == nil {
		return "", "", errors.New("unexpected empty result")
	}
	if len(res.Tenants) == 1 {
		return res.Tenants[0].TenantAlias, res.Tenants[0].Id, nil
	}
	if len(res.Tenants) > 2 {
		return "", "", errors.New("multi-tenancy is not fully supported in m2cp CLI in this version")
	}
	for _, item := range res.Tenants {
		if item.TenantAlias != "mlpa" {
			return item.TenantAlias, item.Id, nil
		}
	}
	return "", "", errors.New("this shouldn't happen")
}

func ProvisionSystemAsset(ctx context.Context, input *structs.SystemAssetProvisionInput) (*structs.SystemAssetProvisionOutput, error) {
	devices := make([]*SystemAssetDeviceProvisionInput, len(input.Devices))
	for i, device := range input.Devices {
		devices[i] = &SystemAssetDeviceProvisionInput{
			Type:               device.Type,
			SerialNumber:       device.SerialNumber,
			McuId:              device.McuId,
			AssetModelName:     device.AssetModelName,
			AssetModelRevision: device.AssetModelRevision,
			AttestationKey:     device.AttestationKey,
		}
	}
	res, err := provisionSystemAsset(ctx, &SystemAssetProvisionInput{
		TenantAlias:          input.TenantAlias,
		SystemAssetModelName: input.SystemAssetModelName,
		SystemAssetName:      input.SystemAssetName,
		Devices:              devices,
	})
	if err != nil {
		return nil, err
	}
	out := &structs.SystemAssetProvisionOutput{
		SystemAsset: structs.SystemAssetProvisionOutputSystem{
			Id:          res.ProvisionSystemAsset.Id,
			AssetName:   tools.MaybeStringToString(res.ProvisionSystemAsset.AssetName, ""),
			SerialNo:    res.ProvisionSystemAsset.SerialNo,
			ChildAssets: make([]*structs.SystemAssetProvisionOutputChildAsset, len(res.ProvisionSystemAsset.ChildAssets)),
		},
	}
	for i, children := range res.ProvisionSystemAsset.ChildAssets {
		out.SystemAsset.ChildAssets[i] = &structs.SystemAssetProvisionOutputChildAsset{
			Id:                 children.Id,
			AssetName:          tools.MaybeStringToString(children.AssetName, ""),
			SerialNo:           children.SerialNo,
			McuId:              children.McuId,
			AttestationKey:     children.AttestationKey,
			AttestationKeySha3: children.AttestationKeySha3,
		}
	}
	return out, nil
}
