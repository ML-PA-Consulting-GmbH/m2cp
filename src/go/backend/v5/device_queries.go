package v5

import (
	"context"
	"encoding/json"
	"fmt"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"strings"
	"time"
)

func makeDeviceFilterCondition(field string, filterInput *StringOperationFilterInput) *DeviceFilterInput {
	switch field {
	case structs.BackendQueryFilterName:
		return &DeviceFilterInput{DeviceName: filterInput}
	}
	panic("unknown field: " + field)
}

func makeDeviceArchFilterInput(arch string) *DeviceFilterInput {

	return &DeviceFilterInput{
		DeviceModelRevision: &DeviceModelRevisionFilterInput{
			DeviceModel: &DeviceModelFilterInput{
				Architecture: &ArchitectureOperationFilterInput{
					Eq: tools.Ptr(Architecture(strings.ToUpper(arch))),
				},
			},
		},
	}
}

func makeDeviceTypeFilterInput(deviceType string) *DeviceFilterInput {
	return &DeviceFilterInput{
		DeviceType: &DeviceTypeFilterInput{
			DeviceTypeName: &StringOperationFilterInput{
				Eq: tools.Ptr(deviceType),
			},
		},
	}
}

func backendQueryFiltersToDeviceFilterAndSort(filters []structs.BackendQueryFilter) (filter *DeviceFilterInput, sort []*DeviceSortInput, err error) {
	conditions := make([]*DeviceFilterInput, 0)
	deviceSort := make([]*DeviceSortInput, 0)
	for _, f := range filters {
		if f.Pattern != nil {
			conditions = append(conditions, makeDeviceFilterCondition(f.Field, patternToFilterInput(*f.Pattern)))
		}
		if len(f.ExactAny) > 0 {
			typesFilter := make([]*DeviceFilterInput, 0)
			for i := range f.ExactAny {
				switch f.Field {
				case structs.BackendQueryFilterArch:
					typesFilter = append(typesFilter, makeDeviceArchFilterInput(f.ExactAny[i]))
				case structs.BackendQueryFilterType:
					typesFilter = append(typesFilter, makeDeviceTypeFilterInput(f.ExactAny[i]))
				}
			}
			conditions = append(conditions, &DeviceFilterInput{Or: typesFilter})
		}
		if f.After != nil {
			conditions = append(conditions, &DeviceFilterInput{
				DeviceStatus: &DeviceStatusFilterInput{
					LastAppstoreActivity: &ComparableNullableOfDateTimeOperationFilterInput{
						Gte: tools.Ptr(f.After.Format("2006-01-02")),
					},
				},
			})
		}
		if f.Before != nil {
			conditions = append(conditions, &DeviceFilterInput{
				DeviceStatus: &DeviceStatusFilterInput{
					LastAppstoreActivity: &ComparableNullableOfDateTimeOperationFilterInput{
						Lte: tools.Ptr(f.Before.Format("2006-01-02")),
					},
				},
			})
		}
		if f.Sort != nil {
			sortDirection := SortEnumType(*f.Sort)
			switch f.Field {
			case structs.BackendQueryFilterId:
				deviceSort = append(deviceSort, &DeviceSortInput{Id: &sortDirection})
			case structs.BackendQueryFilterOSSerial:
				deviceSort = append(deviceSort, &DeviceSortInput{SerialNumber: &sortDirection})
			case structs.BackendQueryFilterDeviceLastStoreActivity:
				deviceSort = append(deviceSort, &DeviceSortInput{DeviceStatus: &DeviceStatusSortInput{LastAppstoreActivity: &sortDirection}})
			case structs.BackendQueryFilterName:
				deviceSort = append(deviceSort, &DeviceSortInput{DeviceName: &sortDirection})
			default:
				return nil, nil, fmt.Errorf("invalid sort key '%s'", f.Field)
			}
		}
	}

	return &DeviceFilterInput{
		And: conditions,
	}, deviceSort, nil
}

func GetDeviceList(ctx context.Context, filters []structs.BackendQueryFilter, take *int, skip *int) (devices []structs.Device, hasNextPage bool, err error) {
	filter, sort, err := backendQueryFiltersToDeviceFilterAndSort(filters)
	if err != nil {
		return nil, false, err
	}
	res, err := getDeviceList(ctx, filter, sort, take, skip)
	if err != nil {
		return nil, false, err
	}
	devices = make([]structs.Device, len(res.Devices.Items))
	for i, resDevice := range res.Devices.Items {
		device := structs.Device{
			DeviceId:     resDevice.Id,
			DeviceName:   resDevice.DeviceName,
			DeviceSerial: resDevice.SerialNumber,
			Description:  resDevice.Description,
		}
		if resDevice.DeviceStatus != nil {
			device.LastAppstoreActivity = tools.MaybeStringToMaybeTime(resDevice.DeviceStatus.LastAppstoreActivity, time.RFC3339)
			device.LastMessagingActivity = tools.MaybeStringToMaybeTime(resDevice.DeviceStatus.LastMessagingActivity, time.RFC3339)
			device.IsOnline = &resDevice.DeviceStatus.IsOnline
		}
		if resDevice.DeviceSnap != nil {
			// DeviceSnap --> it's an Edge Device
			device.UplinkMode = tools.Ptr(string(resDevice.DeviceSnap.UplinkMode))
		}
		if resDevice.DeploymentGroup != nil {
			device.DeploymentGroup = &structs.DeploymentGroup{
				Name: resDevice.DeploymentGroup.Name,
			}
		}
		if resDevice.DeviceModelRevision != nil {
			device.DeviceModelRevision = &structs.DeviceModelRevision{
				Name:     resDevice.DeviceModelRevision.DeviceModel.ModelName,
				Revision: tools.MaybeIntToInt(resDevice.DeviceModelRevision.Revision, -1),
			}
		}
		if resDevice.DeviceType != nil {
			var ok bool
			if device.DeviceType, ok = structs.DeviceTypeAliases[resDevice.DeviceType.DeviceTypeName]; !ok {
				device.DeviceType = structs.DeviceTypeUnknown
			}
		}

		if resDevice.DeviceModelRevision != nil && resDevice.DeviceModelRevision.DeviceModel != nil {
			device.DeviceArchitecture = string(resDevice.DeviceModelRevision.DeviceModel.Architecture)
		}

		devices[i] = device
	}
	if res.Devices.PageInfo != nil {
		hasNextPage = res.Devices.PageInfo.HasNextPage
	}
	return devices, hasNextPage, nil
}

func GetDeviceInfoById(ctx context.Context, id string) (*structs.Device, error) {
	// Query 1: Basic device info
	resBasic, err := getDeviceInfoBasic(ctx, id)
	if err != nil {
		return nil, err
	}
	if resBasic.Device == nil {
		return nil, fmt.Errorf("Device %s not found", id)
	}

	deviceInfo := structs.Device{
		DeviceId:           resBasic.Device.Id,
		DeviceSerial:       resBasic.Device.SerialNumber,
		DeviceName:         resBasic.Device.DeviceName,
		Description:        resBasic.Device.Description,
		DeviceArchitecture: string(resBasic.Device.DeviceModelRevision.DeviceModel.Architecture),
		DeviceModelRevision: &structs.DeviceModelRevision{
			Id:           resBasic.Device.DeviceModelRevision.Id,
			Name:         resBasic.Device.DeviceModelRevision.DeviceModel.ModelName,
			Architecture: string(resBasic.Device.DeviceModelRevision.DeviceModel.Architecture),
			Type:         resBasic.Device.DeviceModelRevision.DeviceModel.DeviceType.DeviceTypeName,
			Revision:     tools.MaybeIntToInt(resBasic.Device.DeviceModelRevision.Revision, 0),
			SystemApps:   []structs.App{},
		},
		DeploymentGroup:      nil,
		DeviceInstallStates:  []structs.DeploymentGroupAppRevision{},
		DevicePendingActions: []structs.DevicePendingAction{},
		DeviceAssetId:        resBasic.Device.AssetId,
	}

	if resBasic.Device.DeviceSnap != nil {
		deviceInfo.DeviceAttestationKey = tools.StrPtr(resBasic.Device.DeviceSnap.DeviceTpmEkPublicKey)
	}

	if resBasic.Device.AssetId != nil {
		deviceInfo.DeviceAssetAccess = tools.BoolPtr(resBasic.Device.Asset != nil && resBasic.Device.Asset.Id == *resBasic.Device.AssetId)
	}

	// populate status
	status := resBasic.Device.DeviceStatus
	if status != nil {
		deviceInfo.LastUptime = &status.LastUptime
		if status.LastMessagingActivity != nil {
			if t, err := time.Parse(time.RFC3339, *status.LastMessagingActivity); err != nil {
				return nil, err
			} else {
				deviceInfo.LastMessagingActivity = &t
			}
		}
		if status.LastAppstoreActivity != nil {
			if t, err := time.Parse(time.RFC3339, *status.LastAppstoreActivity); err != nil {
				return nil, err
			} else {
				deviceInfo.LastAppstoreActivity = &t
			}
			deviceInfo.LastHubEndpoint = resBasic.Device.DeviceStatus.LastHubEndpoint
			if resBasic.Device.DeviceSnap != nil {
				deviceInfo.UplinkMode = tools.StrPtr(string(resBasic.Device.DeviceSnap.UplinkMode))
			}
		}
		if status.LastUplinkSignalContent != nil {
			lastUplinkSignalContent := &structs.DeviceUplinkSignalContent{}
			if err = json.Unmarshal([]byte(*status.LastUplinkSignalContent), lastUplinkSignalContent); err == nil {
				deviceInfo.LastUplinkSignalContent = lastUplinkSignalContent
			}
		}
	}

	// populate system app
	for _, bridge := range resBasic.Device.DeviceModelRevision.DeviceModelRevisionBridgeApps {
		deviceInfo.DeviceModelRevision.SystemApps = append(deviceInfo.DeviceModelRevision.SystemApps, structs.App{
			Id:          bridge.App.Id,
			Name:        bridge.App.AppName,
			Description: bridge.App.Description,
		})
	}

	// populate pending actions
	for _, action := range resBasic.Device.PendingActions.Items {
		deviceInfo.DevicePendingActions = append(deviceInfo.DevicePendingActions, structs.DevicePendingAction{
			AppName:           action.AppName,
			DeltaFileSize:     action.DeltaFileSize,
			DownloadSize:      action.DownloadSize,
			InstalledRevision: action.InstalledRevision,
			InstalledVersion:  action.InstalledVersion,
			IsCoreApp:         action.IsCoreApp,
			PendingActionName: action.PendingActionName,
			TargetRevision:    action.TargetRevision,
			TargetVersion:     action.TargetVersion,
		})
	}

	// populate Edge Device acting as uplink for this RTD
	if resBasic.Device.ConnectedDevice != nil {
		deviceInfo.LastEdgeDevice = &structs.Device{
			DeviceId:     resBasic.Device.ConnectedDevice.Id,
			DeviceName:   resBasic.Device.ConnectedDevice.DeviceName,
			DeviceSerial: resBasic.Device.ConnectedDevice.SerialNumber,
		}
	}

	// populate RTDs connected downlink from this ED
	if resBasic.Device.ConnectedDevices != nil && len(resBasic.Device.ConnectedDevices) > 0 {
		deviceInfo.LastRealTimeDevices = []structs.Device{}
		for _, device := range resBasic.Device.ConnectedDevices {
			deviceInfo.LastRealTimeDevices = append(deviceInfo.LastRealTimeDevices, structs.Device{
				DeviceId:           device.Id,
				DeviceSerial:       device.SerialNumber,
				DeviceName:         device.DeviceName,
				DeviceArchitecture: string(device.DeviceModelRevision.DeviceModel.Architecture),
				DeviceModelRevision: &structs.DeviceModelRevision{
					Name:         device.DeviceModelRevision.DeviceModel.ModelName,
					Architecture: string(device.DeviceModelRevision.DeviceModel.Architecture),
					Revision:     *device.DeviceModelRevision.Revision,
				},
			})
		}
	}

	// Query 2: Deployment group
	resDeploymentGroup, err := getDeviceDeploymentGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	if resDeploymentGroup.Device != nil && resDeploymentGroup.Device.DeploymentGroup != nil {
		owner, err := GetUserById(ctx, resDeploymentGroup.Device.DeploymentGroup.OwnerUserId)
		if err != nil {
			owner = &structs.User{
				Id:   resDeploymentGroup.Device.DeploymentGroup.OwnerUserId,
				Name: fmt.Sprintf("unknown user %s", resDeploymentGroup.Device.DeploymentGroup.OwnerUserId),
			}
		}
		deviceInfo.DeploymentGroup = &structs.DeploymentGroup{
			Id:          resDeploymentGroup.Device.DeploymentGroup.Id,
			Name:        resDeploymentGroup.Device.DeploymentGroup.Name,
			Description: resDeploymentGroup.Device.DeploymentGroup.Description,
			Owner:       *owner,
			CoOwners:    []structs.User{},
		}

		for _, coAdmin := range resDeploymentGroup.Device.DeploymentGroup.DeploymentGroupAdministrators {
			coAdminDetails, err := GetUserById(ctx, coAdmin.UserId)
			if err != nil {
				coAdminDetails = &structs.User{
					Id:   resDeploymentGroup.Device.DeploymentGroup.OwnerUserId,
					Name: fmt.Sprintf("unknown user %s", coAdmin.UserId),
				}
			}
			deviceInfo.DeploymentGroup.CoOwners = append(deviceInfo.DeploymentGroup.CoOwners, structs.User{
				Id:    coAdminDetails.Id,
				Name:  coAdminDetails.Name,
				Email: coAdminDetails.Email,
			})
		}
	}

	// Query 3: Install states
	resInstallStates, err := getDeviceInstallStates(ctx, id)
	if err != nil {
		return nil, err
	}
	if resInstallStates.Device != nil {
		for _, bridge := range resInstallStates.Device.DeviceInstallStates {
			deviceInfo.DeviceInstallStates = append(deviceInfo.DeviceInstallStates, structs.DeploymentGroupAppRevision{
				Id:                      bridge.AppRevision.Id,
				AppId:                   bridge.AppRevision.AppId,
				AppName:                 bridge.AppRevision.App.AppName,
				AppDescription:          bridge.AppRevision.App.Description,
				AppRevisionId:           bridge.AppRevision.Id,
				AppRevision:             bridge.AppRevision.Revision,
				AppVersion:              bridge.AppRevision.Version,
				AppRevisionDownloadSize: bridge.AppRevision.DownloadSize,
			})
		}
	}

	return &deviceInfo, nil
}

func GetDeviceSerialByName(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("&name is required")
	}

	result, err := getDeviceSerialByName(ctx, &name)
	if err != nil {
		return "", err
	}

	if result.Data == nil || result.Data.Items == nil || len(result.Data.Items) == 0 {
		return "", fmt.Errorf("device not found")
	}

	return result.Data.Items[0].DeviceSerial, nil
}

func GetDeviceIdByIdOrNameOrSerial(ctx context.Context, nameOrSerial string) (string, error) {
	if nameOrSerial == "" {
		return "", fmt.Errorf("name or serial is required")
	}

	id := ""

	if tools.IsValidUuid(nameOrSerial) {
		if result, err := getDeviceInfoByIdTiny(ctx, nameOrSerial); err == nil && result != nil && result.Device != nil {
			return result.Device.Id, nil
		}

		result, err := getDeviceIdBySerial(ctx, &nameOrSerial)

		if err != nil {
			return "", err
		}

		if result.Data != nil && result.Data.Items != nil && len(result.Data.Items) > 0 {
			id = result.Data.Items[0].Id
		}
	} else {
		result, err := getDeviceIdByName(ctx, &nameOrSerial)
		if err != nil {
			return "", err
		}

		if result.Data != nil && result.Data.Items != nil && len(result.Data.Items) > 0 {
			id = result.Data.Items[0].Id
		}
	}

	if id == "" {
		return "", fmt.Errorf("device not found")
	}

	return id, nil
}

func GetDeviceWithExtendedFleetInfoBySerial(ctx context.Context, deviceSerial string) (*GetDevicesWithExtendedFleetInfoBySerialResultEdgeDeviceCollectionSegmentItemsEdgeDevice, error) {
	if deviceSerial == "" {
		return nil, fmt.Errorf("deviceSerial is required")
	}

	devices, err := GetDevicesWithExtendedFleetInfoBySerial(ctx, &deviceSerial)
	if err != nil {
		return nil, err
	}
	if len(devices.Result.Items) != 1 {
		return nil, fmt.Errorf("device not found")
	}

	return devices.Result.Items[0], nil
}

func GetCoreAppIdsByDeviceId(ctx context.Context, deviceId string) ([]string, error) {
	if deviceId == "" {
		return nil, fmt.Errorf("deviceId is required")
	}

	response, err := getCoreAppIdsByDeviceId(ctx, deviceId)
	if err != nil {
		return nil, err
	}

	result := []string{}

	if response.Data != nil && response.Data.Items != nil {
		for _, item := range response.Data.Items {
			result = append(result, item.SnapDeclarationId)
		}
	}

	return result, nil

}
