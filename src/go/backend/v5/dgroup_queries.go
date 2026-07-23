package v5

import (
	"context"
	"fmt"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"time"
)

func GetDeploymentGroupByNameOrId(ctx context.Context, arg string) (*structs.DeploymentGroup, error) {
	if !tools.IsValidUuid(arg) {
		res, err := getDeploymentGroupByName(ctx, &arg)
		if err != nil {
			return nil, err
		}
		if res.DeploymentGroups == nil || res.DeploymentGroups.Items == nil || len(res.DeploymentGroups.Items) == 0 {
			return nil, fmt.Errorf("Deployment Group '%s' not found", arg)
		}
		arg = res.DeploymentGroups.Items[0].Id
	}

	res, err := getDeploymentGroupById(ctx, &arg)
	if err != nil {
		return nil, err
	}
	if res.DeploymentGroups == nil || res.DeploymentGroups.Items == nil || len(res.DeploymentGroups.Items) == 0 {
		return nil, fmt.Errorf("Deployment Group '%s' not found", arg)
	}
	dgroup := res.DeploymentGroups.Items[0]

	dgroupsInfo := &structs.DeploymentGroup{
		Id:           dgroup.Id,
		Name:         dgroup.Name,
		Architecture: string(dgroup.DeviceModelRevision.DeviceModel.Architecture),
		Devices:      []structs.Device{},
		DeviceType:   dgroup.DeviceModelRevision.DeviceModel.DeviceType.DeviceTypeName,
		AutoUpdate:   dgroup.AutoUpdateMode.Name,
		ModelRevision: &structs.DeviceModelRevision{
			Id:           dgroup.DeviceModelRevision.Id,
			Name:         dgroup.DeviceModelRevision.DeviceModel.ModelName,
			Type:         dgroup.DeviceModelRevision.DeviceModel.DeviceType.DeviceTypeName,
			Architecture: string(dgroup.DeviceModelRevision.DeviceModel.Architecture),
			Revision:     tools.MaybeIntToInt(dgroup.DeviceModelRevision.Revision, -1),
		},
		CoOwners:            []structs.User{},
		PendingActionsTotal: dgroup.PendingActionSummary.TotalPendingActionCount,
		IsDeltaUpdateOnly:   dgroup.IsDeltaUpdateOnly,
	}

	// resolve owner
	owner, err := GetUserById(ctx, dgroup.OwnerUserId)
	if err != nil {
		return nil, err
	}
	dgroupsInfo.Owner = structs.User{
		Id:    owner.Id,
		Name:  owner.Name,
		Email: owner.Email,
	}

	// resolve co-admins
	for _, user := range dgroup.DeploymentGroupAdministrators {
		var coOwner *structs.User
		coOwner, err = GetUserById(ctx, user.UserId)
		if err != nil {
			return nil, err
		}
		dgroupsInfo.CoOwners = append(dgroupsInfo.CoOwners, *coOwner)
	}

	for _, device := range dgroup.Devices {
		deviceInfo := structs.Device{
			DeviceId:     device.Id,
			DeviceSerial: device.SerialNumber,
			DeviceName:   device.DeviceName,
		}
		if device.DeviceStatus != nil {
			if device.DeviceStatus.LastAppstoreActivity != nil {
				t, _ := time.Parse(time.RFC3339, *device.DeviceStatus.LastAppstoreActivity)
				deviceInfo.LastAppstoreActivity = &t
			}
			if device.DeviceStatus.LastMessagingActivity != nil {
				t, _ := time.Parse(time.RFC3339, *device.DeviceStatus.LastMessagingActivity)
				deviceInfo.LastMessagingActivity = &t
			}
			deviceInfo.LastUptime = &device.DeviceStatus.LastUptime
		}

		dgroupsInfo.Devices = append(dgroupsInfo.Devices, deviceInfo)
	}

	for _, app := range dgroup.DeploymentGroupBridgeAppRevisions {
		dgroupsInfo.AppRevisions = append(dgroupsInfo.AppRevisions, structs.DeploymentGroupAppRevision{
			AppId:                    app.AppRevision.App.Id,
			AppRevisionId:            app.AppRevision.Id,
			AppName:                  app.AppRevision.App.AppName,
			AppRevision:              app.AppRevision.Revision,
			AppVersion:               app.AppRevision.Version,
			AppDescription:           app.AppRevision.App.Description,
			AppRating:                app.AppRevision.AppStatus.Name,
			AppRevisionUploadMessage: app.AppRevision.UploadMessage,
			IsSystemApp:              &app.IsCoreApp,
		})

	}

	return dgroupsInfo, nil
}
