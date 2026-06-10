package legacy

import (
	"context"
	"errors"
	"fmt"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"strings"
)

func SetAssetSystem(ctx context.Context, deviceAssetId string, systemAssetId *string) error {
	_, err := setAssetSystem(ctx, deviceAssetId, systemAssetId, tools.BoolPtr(systemAssetId == nil))
	return err
}

func CreateAsset(ctx context.Context, modelId string, hwSerial string, assetName string, assetDescription *string, attestationKey *string) (id string, err error) {
	_, ownerId, err := GetAssetsOwner(ctx)
	if err != nil {
		return "", err
	}
	res, err := createAsset(ctx, modelId, hwSerial, assetName, assetDescription, attestationKey, ownerId)
	if err != nil {
		return "", err
	}
	if res.CreateAssets == nil || len(res.CreateAssets) == 0 {
		return "", fmt.Errorf("query failed with empty result")
	}
	return res.CreateAssets[0].Id, nil
}

func CreateAssetModel(ctx context.Context, modelTypeId string, name string, description *string) (id string, err error) {
	res, err := createAssetModel(ctx, modelTypeId, name, description)
	if err != nil {
		return "", err
	}
	if res == nil || res.CreateAssetModels == nil {
		return "", fmt.Errorf("query failed with nil result")
	}
	return res.CreateAssetModels[0].Id, nil
}

func GetAssetModels(ctx context.Context, types []string) ([]structs.AssetModel, error) {
	res, err := assetModels(ctx, types)
	if err != nil {
		return nil, err
	}
	if res == nil || res.AssetModels == nil || res.AssetModels.Items == nil {
		return nil, fmt.Errorf("query failed with nil result")
	}
	models := make([]structs.AssetModel, len(res.AssetModels.Items))
	for i, assetModel := range res.AssetModels.Items {
		models[i] = structs.AssetModel{
			Id:          assetModel.Id,
			Name:        assetModel.AssetModelName,
			Description: assetModel.Description,
			Type:        structs.AssetModelTypeIdToType(assetModel.AssetTypeId),
			Revision:    assetModel.Revision,
		}
	}
	return models, nil
}

func GetAssetModelById(ctx context.Context, assetModelId string) (*structs.AssetModel, error) {
	res, err := assetModelById(ctx, assetModelId)
	if err != nil {
		return nil, err
	}
	if res == nil || res.AssetModel == nil {
		return nil, fmt.Errorf("query failed with nil result")
	}
	return &structs.AssetModel{
		Id:          res.AssetModel.Id,
		Name:        res.AssetModel.AssetModelName,
		Description: res.AssetModel.Description,
		Type:        structs.AssetModelTypeIdToType(res.AssetModel.AssetTypeId),
		Revision:    res.AssetModel.Revision,
		TenantId:    res.AssetModel.Tenant.Id,
		TenantName:  res.AssetModel.Tenant.Alias,
	}, nil
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
			Name:                   asset.AssetName,
			Serial:                 asset.SerialNo,
			AssetModelName:         asset.AssetModel.AssetModelName,
			AssetModelTypeName:     asset.AssetModel.AssetType.AssetTypeName,
			AssetModelTypeCategory: string(asset.AssetModel.AssetType.AssetTypeCategory),
			DeviceId:               asset.DeviceId,
		}
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
		Name:                   asset.AssetName,
		Serial:                 asset.SerialNo,
		AssetModelName:         asset.AssetModel.AssetModelName,
		AssetModelId:           asset.AssetModel.Id,
		AssetType:              tools.StrPtr(structs.AssetModelTypeIdToType(asset.AssetModel.AssetTypeId)),
		AssetModelTypeName:     asset.AssetModel.AssetType.AssetTypeName,
		AssetModelTypeCategory: string(asset.AssetModel.AssetType.AssetTypeCategory),
		TenantId:               asset.Tenant.Id,
		TenantName:             asset.Tenant.Alias,
		DeviceId:               asset.DeviceId,
	}
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
					Components:             []structs.Asset{},
					IsSystemOwned:          tools.Ptr(resItem.IsSystemOwned),
					Name:                   resItem.AssetName,
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

func ProvisionSystemAsset(ctx context.Context, input *structs.SystemAssetProvisionInput) (*structs.SystemAssetProvisionOutput, error) {
	devices := make([]*SystemAssetDeviceProvisionInput, len(input.Devices))
	for i, device := range input.Devices {
		devices[i] = &SystemAssetDeviceProvisionInput{
			Type:               device.Type,
			SerialNumber:       device.SerialNumber,
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
			AssetName:   res.ProvisionSystemAsset.AssetName,
			SerialNo:    res.ProvisionSystemAsset.SerialNo,
			ChildAssets: make([]*structs.SystemAssetProvisionOutputChildAsset, len(res.ProvisionSystemAsset.ChildAssets)),
		},
	}
	for i, children := range res.ProvisionSystemAsset.ChildAssets {
		out.SystemAsset.ChildAssets[i] = &structs.SystemAssetProvisionOutputChildAsset{
			Id:                 children.Id,
			AssetName:          children.AssetName,
			SerialNo:           children.SerialNo,
			AttestationKey:     children.AttestationKey,
			AttestationKeySha3: children.AttestationKeySha3,
		}
	}
	return out, nil
}
