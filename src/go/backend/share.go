package backend

import (
	"context"
	"fmt"
	v5 "m2cpcli/backend/v5"
)

type (
	ShareAppResponse         = v5.ShareAppResponse
	ShareDeviceModelResponse = v5.ShareDeviceModelResponse
)

// ShareApp makes an app globally shared across all tenants. This is irreversible.
func ShareApp(ctx context.Context, appId string) (*ShareAppResponse, error) {
	if backendMajorVersionInt(ctx) < 5 {
		return nil, fmt.Errorf("sharing apps globally requires backend v5 or later")
	}
	return v5.ShareApp(ctx, appId)
}

// ShareDeviceModel makes a model globally shared across all tenants. This is irreversible.
func ShareDeviceModel(ctx context.Context, deviceModelId string) (*ShareDeviceModelResponse, error) {
	if backendMajorVersionInt(ctx) < 5 {
		return nil, fmt.Errorf("sharing models globally requires backend v5 or later")
	}
	return v5.ShareDeviceModel(ctx, deviceModelId)
}

// GetDeviceModelIdByRevisionId resolves the id of the parent model for a given model revision id.
// Global sharing operates on the model, not a specific revision.
func GetDeviceModelIdByRevisionId(ctx context.Context, revisionId string) (string, error) {
	if backendMajorVersionInt(ctx) < 5 {
		return "", fmt.Errorf("resolving a model's id from its revision requires backend v5 or later")
	}

	result, err := v5.GetDeviceModelIdByRevisionId(ctx, revisionId)
	if err != nil {
		return "", err
	}
	if result.DeviceModelRevision == nil {
		return "", fmt.Errorf("could not find a model revision with id '%s'", revisionId)
	}

	return result.DeviceModelRevision.DeviceModelId, nil
}

// GetAppsGlobalShareStatus returns, for each given app id, whether it is globally shared.
// supported is false if the backend doesn't support global sharing (backend < v5), in which
// case status is nil.
func GetAppsGlobalShareStatus(ctx context.Context, appIds []string) (status map[string]bool, supported bool, err error) {
	if backendMajorVersionInt(ctx) < 5 {
		return nil, false, nil
	}
	if len(appIds) == 0 {
		return map[string]bool{}, true, nil
	}

	take := len(appIds)
	result, err := v5.GetAppsGlobalShareStatus(ctx, appIds, &take)
	if err != nil {
		return nil, true, err
	}

	status = make(map[string]bool, len(appIds))
	if result.Apps != nil {
		for _, item := range result.Apps.Items {
			status[item.Id] = item.IsGloballyShared
		}
	}
	return status, true, nil
}

// GetDeviceModelsGlobalShareStatus returns, for each given model id, whether it is globally shared.
// supported is false if the backend doesn't support global sharing (backend < v5), in which
// case status is nil.
func GetDeviceModelsGlobalShareStatus(ctx context.Context, deviceModelIds []string) (status map[string]bool, supported bool, err error) {
	if backendMajorVersionInt(ctx) < 5 {
		return nil, false, nil
	}
	if len(deviceModelIds) == 0 {
		return map[string]bool{}, true, nil
	}

	take := len(deviceModelIds)
	result, err := v5.GetDeviceModelsGlobalShareStatus(ctx, deviceModelIds, &take)
	if err != nil {
		return nil, true, err
	}

	status = make(map[string]bool, len(deviceModelIds))
	if result.DeviceModels != nil {
		for _, item := range result.DeviceModels.Items {
			status[item.Id] = item.IsGloballyShared
		}
	}
	return status, true, nil
}
