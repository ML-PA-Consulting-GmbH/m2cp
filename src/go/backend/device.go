package backend

import (
	"context"
	"m2cpcli/backend/legacy"
	v5 "m2cpcli/backend/v5"
	"m2cpcli/structs"
)

type (
	DeviceInstallState                                                                                         = legacy.DeviceInstallState
	DeviceModelRevisionFilterInput                                                                             = legacy.DeviceModelRevisionFilterInput
	ComparableGuidOperationFilterInput                                                                         = legacy.ComparableGuidOperationFilterInput
	ComparableNullableOfInt32OperationFilterInput                                                              = legacy.ComparableNullableOfInt32OperationFilterInput
	DeviceModelFilterInput                                                                                     = legacy.DeviceModelFilterInput
	StringOperationFilterInput                                                                                 = legacy.StringOperationFilterInput
	GetDeviceModelRevisionListDeviceModelRevisionsDeviceModelRevisionCollectionSegmentItemsDeviceModelRevision = legacy.GetDeviceModelRevisionListDeviceModelRevisionsDeviceModelRevisionCollectionSegmentItemsDeviceModelRevision
	GetDevicesWithExtendedFleetInfoBySerialResultEdgeDeviceCollectionSegmentItemsEdgeDevice                    = legacy.GetDevicesWithExtendedFleetInfoBySerialResultEdgeDeviceCollectionSegmentItemsEdgeDevice
	RpcResponse                                                                                                = legacy.RpcResponse
	DeviceFilterInput                                                                                          = legacy.DeviceFilterInput
	DeviceSortInput                                                                                            = legacy.DeviceSortInput
	EdgeDeviceInstallStateHistoryFilterInput                                                                   = legacy.EdgeDeviceInstallStateHistoryFilterInput
	SnapRevisionFilterInput                                                                                    = legacy.SnapRevisionFilterInput
	SnapDeclarationFilterInput                                                                                 = legacy.SnapDeclarationFilterInput
	GetEdgeDeviceInstallHistoriesResponse                                                                      = legacy.GetEdgeDeviceInstallHistoriesResponse
	GetEdgeDeviceInstallSnapshotsResponse                                                                      = legacy.GetEdgeDeviceInstallSnapshotsResponse
	EdgeDeviceInstallStateSnapshotFilterInput                                                                  = legacy.EdgeDeviceInstallStateSnapshotFilterInput
	ComparableDateTimeOperationFilterInput                                                                     = legacy.ComparableDateTimeOperationFilterInput
	GetDeviceModelRevisionInfoByIdResponse                                                                     = legacy.GetDeviceModelRevisionInfoByIdResponse
	GetDeviceModelRevisionInfoByIdDeviceModelRevision                                                          = legacy.GetDeviceModelRevisionInfoByIdDeviceModelRevision
	UpdateDeviceModelRevisionResponse                                                                          = legacy.UpdateDeviceModelRevisionResponse
	UplinkMode                                                                                                 = legacy.UplinkMode
	DeviceStatusFilterInput                                                                                    = legacy.DeviceStatusFilterInput
	ComparableNullableOfDateTimeOperationFilterInput                                                           = legacy.ComparableNullableOfDateTimeOperationFilterInput
	DeviceStatusSortInput                                                                                      = legacy.DeviceStatusSortInput
	EdgeDeviceLogQueryInput                                                                                    = legacy.EdgeDeviceLogQueryInput
	GetDeviceLogsResponse                                                                                      = legacy.GetDeviceLogsResponse
)

func GetDevicesInstallStates(ctx context.Context, deviceIds []string) (map[string][]legacy.DeviceInstallState, error) {
	return legacy.GetDevicesInstallStates(ctx, deviceIds)
}

func EdgeDeviceRefresh(ctx context.Context, osSerial string) error {
	return legacy.EdgeDeviceRefresh(ctx, osSerial)
}

func GetDeviceTypeIdByName(deviceTypeName string) (string, bool) {
	return legacy.GetDeviceTypeIdByName(deviceTypeName)
}

func GetDeviceTypeNameById(deviceTypeId string) string {
	return legacy.GetDeviceTypeNameById(deviceTypeId)
}

func TryGetDeviceModelRevisionById(ctx context.Context, modelRevisionId string) (string, bool) {
	return legacy.TryGetDeviceModelRevisionById(ctx, modelRevisionId)
}

func TryGetDeviceModelRevisionByTypeNameRevision(ctx context.Context, deviceTypeName, modelName string, revision int) (string, error) {
	return legacy.TryGetDeviceModelRevisionByTypeNameRevision(ctx, deviceTypeName, modelName, revision)
}

func GetDeviceModelRevisionInfo(ctx context.Context, modelRevisionId string) (*legacy.GetDeviceModelRevisionListDeviceModelRevisionsDeviceModelRevisionCollectionSegmentItemsDeviceModelRevision, error) {
	return legacy.GetDeviceModelRevisionInfo(ctx, modelRevisionId)
}

func GetDeviceInfoById(ctx context.Context, id string) (*structs.Device, error) {
	return legacy.GetDeviceInfoById(ctx, id)
}

func GetDeviceSerialByName(ctx context.Context, name string) (string, error) {
	return legacy.GetDeviceSerialByName(ctx, name)
}

func GetDeviceIdByIdOrNameOrSerial(ctx context.Context, nameOrSerial string) (string, error) {
	return legacy.GetDeviceIdByIdOrNameOrSerial(ctx, nameOrSerial)
}

func GetDeviceWithExtendedFleetInfoBySerial(ctx context.Context, deviceSerial string) (*legacy.GetDevicesWithExtendedFleetInfoBySerialResultEdgeDeviceCollectionSegmentItemsEdgeDevice, error) {
	return legacy.GetDeviceWithExtendedFleetInfoBySerial(ctx, deviceSerial)
}

func GetCoreAppIdsByDeviceId(ctx context.Context, deviceId string) ([]string, error) {
	return legacy.GetCoreAppIdsByDeviceId(ctx, deviceId)
}

func GetEdgeDeviceInstallHistories(ctx context.Context, filter *EdgeDeviceInstallStateHistoryFilterInput, take *int) (*legacy.GetEdgeDeviceInstallHistoriesResponse, error) {
	return legacy.GetEdgeDeviceInstallHistories(ctx, (*legacy.EdgeDeviceInstallStateHistoryFilterInput)(filter), take)
}

func GetDeviceModelRevisionInfoById(ctx context.Context, id string) (*legacy.GetDeviceModelRevisionInfoByIdResponse, error) {
	return legacy.GetDeviceModelRevisionInfoById(ctx, id)
}

func GetDeviceModelRevisionList(ctx context.Context, filter *DeviceModelRevisionFilterInput, take *int, skip *int) (*legacy.GetDeviceModelRevisionListResponse, error) {
	return legacy.GetDeviceModelRevisionList(ctx, (*legacy.DeviceModelRevisionFilterInput)(filter), take, skip)
}

func UpdateDeviceModelRevision(ctx context.Context, id string, isTpmRequired *bool, isPreRegistrationRequired *bool) (*legacy.UpdateDeviceModelRevisionResponse, error) {
	return legacy.UpdateDeviceModelRevision(ctx, id, isTpmRequired, isPreRegistrationRequired)
}

func DeviceRpc(ctx context.Context, node, command string, params map[string]string) (*legacy.RpcResponse, error) {
	return legacy.DeviceRpc(ctx, node, command, params)
}

func GetDeviceList(ctx context.Context, filter []structs.BackendQueryFilter, take *int, skip *int) (devices []structs.Device, hasNextPage bool, err error) {
	if backendMajorVersionInt(ctx) < 5 {
		return legacy.GetDeviceList(ctx, filter, take, skip)
	}
	return v5.GetDeviceList(ctx, filter, take, skip)
}

func GetDevicesOnlineStatus(ctx context.Context, serialNumbers []string) (*legacy.GetDevicesOnlineStatusResponse, error) {
	return legacy.GetDevicesOnlineStatus(ctx, serialNumbers)
}

func GetEdgeDeviceInstallSnapshots(ctx context.Context, filter *EdgeDeviceInstallStateSnapshotFilterInput, take *int) (*legacy.GetEdgeDeviceInstallSnapshotsResponse, error) {
	return legacy.GetEdgeDeviceInstallSnapshots(ctx, (*legacy.EdgeDeviceInstallStateSnapshotFilterInput)(filter), take)
}

func GetDeviceLogs(ctx context.Context, filters *EdgeDeviceLogQueryInput) (*legacy.GetDeviceLogsResponse, error) {
	return legacy.GetDeviceLogs(ctx, (*legacy.EdgeDeviceLogQueryInput)(filters))
}

func ModifyDevice(ctx context.Context, deviceId string, deviceName *string, deviceDescription *string, deviceActivated *bool) (*legacy.ModifyDeviceResponse, error) {
	return legacy.ModifyDevice(ctx, deviceId, deviceName, deviceDescription, deviceActivated)
}

func GetDeviceModelByIdWithAssertion(ctx context.Context, id string) (*legacy.GetDeviceModelByIdWithAssertionResponse, error) {
	return legacy.GetDeviceModelByIdWithAssertion(ctx, id)
}

func SetDeviceDeploymentGroup(ctx context.Context, deviceId string, deploymentGroupId *string) (*legacy.SetDeviceDeploymentGroupResponse, error) {
	return legacy.SetDeviceDeploymentGroup(ctx, deviceId, deploymentGroupId)
}

func UnsetDeviceDeploymentGroup(ctx context.Context, deviceId string) (*legacy.UnsetDeviceDeploymentGroupResponse, error) {
	return legacy.UnsetDeviceDeploymentGroup(ctx, deviceId)
}

func GetEdgeDeviceSshContainerStatus(ctx context.Context, deviceSerial string) (*legacy.GetEdgeDeviceSshContainerStatusResponse, error) {
	return legacy.GetEdgeDeviceSshContainerStatus(ctx, deviceSerial)
}

func DeviceSshUserReset(ctx context.Context, deviceSerial string) (*legacy.DeviceSshUserResetResponse, error) {
	return legacy.DeviceSshUserReset(ctx, deviceSerial)
}

func GetFleetInfoById(ctx context.Context, fleetId string) (*legacy.GetFleetInfoByIdResponse, error) {
	return legacy.GetFleetInfoById(ctx, fleetId)
}

func CreateFleetBridgeSnapRevisions(ctx context.Context, fleetId string, snapRevisionId string) (*legacy.CreateFleetBridgeSnapRevisionsResponse, error) {
	return legacy.CreateFleetBridgeSnapRevisions(ctx, fleetId, snapRevisionId)
}

func UpdateFleetBridgeSnapRevision(ctx context.Context, id string, snapRevisionId string) (*legacy.UpdateFleetBridgeSnapRevisionResponse, error) {
	return legacy.UpdateFleetBridgeSnapRevision(ctx, id, snapRevisionId)
}

func SetDeviceFleet(ctx context.Context, deviceId string, fleetId *string) (*legacy.SetDeviceFleetResponse, error) {
	return legacy.SetDeviceFleet(ctx, deviceId, fleetId)
}

func SetDeviceToDead(ctx context.Context, deviceId string) (*legacy.SetDeviceToDeadResponse, error) {
	return legacy.SetDeviceToDead(ctx, deviceId)
}

func DeleteFleet(ctx context.Context, fleetId string) (*legacy.DeleteFleetResponse, error) {
	return legacy.DeleteFleet(ctx, fleetId)
}

func GetDeviceDataMessages(ctx context.Context, edgeDeviceId string, date string) (*legacy.GetDeviceDataMessagesResponse, error) {
	return legacy.GetDeviceDataMessages(ctx, edgeDeviceId, date)
}
