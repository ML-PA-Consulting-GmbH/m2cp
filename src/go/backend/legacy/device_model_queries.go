package legacy

import (
	"context"
	"fmt"
)

// TryGetDeviceModelRevisionById checks if the given ID corresponds to an existing device model revision and returns the ID
func TryGetDeviceModelRevisionById(ctx context.Context, modelRevisionId string) (string, bool) {
	// Create filter for the specific ID
	filter := &DeviceModelRevisionFilterInput{
		Id: &ComparableGuidOperationFilterInput{
			Eq: &modelRevisionId,
		},
	}

	take := 1
	res, err := GetDeviceModelRevisionList(ctx, filter, &take, nil)
	if err != nil {
		return "", false
	}

	if res.DeviceModelRevisions == nil || res.DeviceModelRevisions.Items == nil || len(res.DeviceModelRevisions.Items) == 0 {
		return "", false
	}

	return res.DeviceModelRevisions.Items[0].Id, true
}

// TryGetDeviceModelRevisionByTypeNameRevision looks up a device model revision by device type, model name, and revision number
func TryGetDeviceModelRevisionByTypeNameRevision(ctx context.Context, deviceTypeName, modelName string, revision int) (string, error) {
	// Get device type ID from name/alias
	deviceTypeId, found := GetDeviceTypeIdByName(deviceTypeName)
	if !found {
		return "", fmt.Errorf("unknown device type '%s'", deviceTypeName)
	}

	// WORKAROUND: The genqlient-generated filter structs (e.g. StringOperationFilterInput,
	// ComparableGuidOperationFilterInput) do not have `omitempty` on their JSON tags.
	// As a result, all unset filter fields serialize as explicit `null` values in the request body.
	// Our assumption is that the backend treats these null sibling fields as wildcard conditions,
	// which would explain why filters are ignored and all records are returned — but the root cause
	// has not been fully confirmed.
	// Until the issue is understood and properly fixed (e.g. by adding omitempty to the schema or
	// generator config), model name is used as the server-side filter (more selective, as there are
	// typically more models than revisions per model), and revision / device type filtering is
	// applied client-side on the returned results.
	filter := &DeviceModelRevisionFilterInput{
		DeviceModel: &DeviceModelFilterInput{
			ModelName: &StringOperationFilterInput{
				Eq: &modelName,
			},
		},
	}

	take := 100
	res, err := GetDeviceModelRevisionList(ctx, filter, &take, nil)
	if err != nil {
		return "", fmt.Errorf("failed to query device model revisions: %w", err)
	}

	if res.DeviceModelRevisions == nil || res.DeviceModelRevisions.Items == nil {
		return "", fmt.Errorf("no device model revisions found for type='%s', name='%s', revision=%d", deviceTypeName, modelName, revision)
	}

	// Client-side filter by revision and device type ID
	var matched []string
	for _, item := range res.DeviceModelRevisions.Items {
		if item.DeviceModel.DeviceTypeId == deviceTypeId && item.Revision != nil && *item.Revision == revision {
			matched = append(matched, item.Id)
		}
	}

	if len(matched) == 0 {
		return "", fmt.Errorf("no device model revision found for type='%s', name='%s', revision=%d", deviceTypeName, modelName, revision)
	}

	if len(matched) > 1 {
		return "", fmt.Errorf("multiple device model revisions found for type='%s', name='%s', revision=%d", deviceTypeName, modelName, revision)
	}

	return matched[0], nil
}

// GetDeviceModelRevisionInfo retrieves detailed information about a device model revision
func GetDeviceModelRevisionInfo(ctx context.Context, modelRevisionId string) (*GetDeviceModelRevisionListDeviceModelRevisionsDeviceModelRevisionCollectionSegmentItemsDeviceModelRevision, error) {
	filter := &DeviceModelRevisionFilterInput{
		Id: &ComparableGuidOperationFilterInput{
			Eq: &modelRevisionId,
		},
	}

	take := 1
	res, err := GetDeviceModelRevisionList(ctx, filter, &take, nil)
	if err != nil {
		return nil, err
	}

	if res.DeviceModelRevisions == nil || res.DeviceModelRevisions.Items == nil || len(res.DeviceModelRevisions.Items) == 0 {
		return nil, fmt.Errorf("device model revision '%s' not found", modelRevisionId)
	}

	return res.DeviceModelRevisions.Items[0], nil
}
