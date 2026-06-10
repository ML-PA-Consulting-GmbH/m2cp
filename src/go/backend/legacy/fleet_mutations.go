package legacy

import (
	"context"
	"fmt"
)

func CreateFleetWithAutoUpdate(ctx context.Context, fleetName, fleetDescription string, fleetIsDeltaUpdateOnly bool,
	modelType, modelName string, modelRevision int, autoUpdateMode string) (*CreateFleetsFleet, error) {
	if fleetName == "" {
		return nil, fmt.Errorf("fleet name is required")
	}
	if modelType == "" {
		return nil, fmt.Errorf("model type is required")
	}
	if modelName == "" {
		return nil, fmt.Errorf("model name is required")
	}
	if modelRevision == 0 {
		return nil, fmt.Errorf("model revision is required")
	}
	if autoUpdateMode == "" {
		return nil, fmt.Errorf("auto update mode is required")
	}

	edgeDeviceModel, err := GetDeviceModelIdByTypeNameRevision(ctx, modelType, modelName, modelRevision)
	if err != nil {
		return nil, err
	}
	if len(edgeDeviceModel.EdgeDeviceModels.Items) == 0 {
		return nil, fmt.Errorf("model not found")
	}

	autoUpdateModeId, found := GetAutoUpdateModeIdByName(autoUpdateMode)
	if !found {
		return nil, fmt.Errorf("unknown auto update mode '%s'", autoUpdateMode)
	}

	isDeltaUpdateOnly := false
	createdFleet, err := CreateFleets(ctx, fleetName, edgeDeviceModel.EdgeDeviceModels.Items[0].Id, &fleetDescription, &isDeltaUpdateOnly, &autoUpdateModeId)
	if err != nil {
		return nil, err
	}
	if len(createdFleet.Fleet) != 1 {
		return nil, fmt.Errorf("created not the expected number of fleets, expected 1, got %d", len(createdFleet.Fleet))
	}

	return createdFleet.Fleet[0], nil
}

func CreateFleet(ctx context.Context, fleetName, fleetDescription string, fleetIsDeltaUpdateOnly bool,
	modelType, modelName string, modelRevision int) (*CreateFleetsFleet, error) {
	return CreateFleetWithAutoUpdate(ctx, fleetName, fleetDescription, fleetIsDeltaUpdateOnly, modelType, modelName, modelRevision, "off")
}
