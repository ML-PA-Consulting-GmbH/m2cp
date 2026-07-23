package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	"m2cpcli/backend/v5"
	"m2cpcli/structs"
)

const (
	EdgeAutoUpdateModeId   = legacy.EdgeAutoUpdateModeId
	OffAutoUpdateModeId    = legacy.OffAutoUpdateModeId
	StableAutoUpdateModeId = legacy.StableAutoUpdateModeId
)

type (
	CreateFleetsFleet                    = legacy.CreateFleetsFleet
	DeploymentGroupCloneInput            = legacy.DeploymentGroupCloneInput
	DeploymentGroupCreateInput           = legacy.DeploymentGroupCreateInput
	DeploymentGroupCreateFromDeviceInput = legacy.DeploymentGroupCreateFromDeviceInput
	GetDeploymentGroupLogBooksResponse   = legacy.GetDeploymentGroupLogBooksResponse
)

func GetAutoUpdateModeIdByName(autoUpdateModeName string) (string, bool) {
	return legacy.GetAutoUpdateModeIdByName(autoUpdateModeName)
}

func GetAutoUpdateModeNameById(autoUpdateModeId string) string {
	return legacy.GetAutoUpdateModeNameById(autoUpdateModeId)
}

func GetDeploymentGroupByNameOrId(ctx context.Context, arg string) (*structs.DeploymentGroup, error) {
	return v5.GetDeploymentGroupByNameOrId(ctx, arg)
}

func GetDeploymentGroupAppRevision(ctx context.Context, dgroupId, appId string) (*structs.DeploymentGroupAppRevision, error) {
	return legacy.GetDeploymentGroupAppRevision(ctx, dgroupId, appId)
}

func RemoveAppFromDeploymentGroup(ctx context.Context, dgroupId, appId string) (bridgeId string, err error) {
	return legacy.RemoveAppFromDeploymentGroup(ctx, dgroupId, appId)
}

func AddAppRevisionToDeploymentGroup(ctx context.Context, dgroupId, appId, appRevisionId string, modifyExisting bool) (bridgeId string, err error) {
	return legacy.AddAppRevisionToDeploymentGroup(ctx, dgroupId, appId, appRevisionId, modifyExisting)
}

func CreateFleetWithAutoUpdate(ctx context.Context, fleetName, fleetDescription string, fleetIsDeltaUpdateOnly bool,
	modelType, modelName string, modelRevision int, autoUpdateMode string) (*legacy.CreateFleetsFleet, error) {
	return legacy.CreateFleetWithAutoUpdate(ctx, fleetName, fleetDescription, fleetIsDeltaUpdateOnly, modelType, modelName, modelRevision, autoUpdateMode)
}

func CreateFleet(ctx context.Context, fleetName, fleetDescription string, fleetIsDeltaUpdateOnly bool,
	modelType, modelName string, modelRevision int) (*legacy.CreateFleetsFleet, error) {
	return legacy.CreateFleet(ctx, fleetName, fleetDescription, fleetIsDeltaUpdateOnly, modelType, modelName, modelRevision)
}

func GetDeploymentGroupIdByName(ctx context.Context, name string) (*legacy.GetDeploymentGroupIdByNameResponse, error) {
	return legacy.GetDeploymentGroupIdByName(ctx, name)
}

func DeploymentGroupClone(ctx context.Context, input *DeploymentGroupCloneInput) (*legacy.DeploymentGroupCloneResponse, error) {
	return legacy.DeploymentGroupClone(ctx, (*legacy.DeploymentGroupCloneInput)(input))
}

func CreateDeploymentGroups(ctx context.Context, input []*DeploymentGroupCreateInput) (*legacy.CreateDeploymentGroupsResponse, error) {
	legacyInput := make([]*legacy.DeploymentGroupCreateInput, len(input))
	for i, v := range input {
		legacyInput[i] = (*legacy.DeploymentGroupCreateInput)(v)
	}
	return legacy.CreateDeploymentGroups(ctx, legacyInput)
}

func CreateDeploymentGroupFromDevice(ctx context.Context, input *DeploymentGroupCreateFromDeviceInput) (*legacy.CreateDeploymentGroupFromDeviceResponse, error) {
	return legacy.CreateDeploymentGroupFromDevice(ctx, (*legacy.DeploymentGroupCreateFromDeviceInput)(input))
}

func DeleteDeploymentGroups(ctx context.Context, ids []string) (*legacy.DeleteDeploymentGroupsResponse, error) {
	return legacy.DeleteDeploymentGroups(ctx, ids)
}

func GetDeploymentGroupLogBooks(ctx context.Context, id string, fromDateTime string, toDateTime string) (*legacy.GetDeploymentGroupLogBooksResponse, error) {
	return legacy.GetDeploymentGroupLogBooks(ctx, id, fromDateTime, toDateTime)
}

func GetFleetIdByName(ctx context.Context, name string) (*legacy.GetFleetIdByNameResponse, error) {
	return legacy.GetFleetIdByName(ctx, name)
}

func UpdateDeploymentGroup(ctx context.Context, id string, name *string, description *string, autoUpdateModeId *string, isDeltaUpdateOnly *bool) (*v5.UpdateDeploymentGroupResponse, error) {
	return v5.UpdateDeploymentGroup(ctx, id, name, description, autoUpdateModeId, isDeltaUpdateOnly)
}
