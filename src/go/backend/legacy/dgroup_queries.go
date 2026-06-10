package legacy

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

func GetDeploymentGroupAppRevision(ctx context.Context, dgroupId, appId string) (*structs.DeploymentGroupAppRevision, error) {
	res, err := GetDeploymentGroupBridgeAppRevision(ctx, dgroupId, appId)
	if err != nil {
		return nil, err
	}
	if res == nil || res.DeploymentGroupBridgeAppRevisions.Items == nil || len(res.DeploymentGroupBridgeAppRevisions.Items) == 0 {
		return nil, fmt.Errorf("Deployment Group bridge '%s'-'%s' not found", dgroupId, appId)
	}
	item := res.DeploymentGroupBridgeAppRevisions.Items[0]
	return &structs.DeploymentGroupAppRevision{
		AppId:         item.AppRevision.App.Id,
		AppName:       item.AppRevision.App.AppName,
		AppRevisionId: item.AppRevision.Id,
		AppRevision:   item.AppRevision.Revision,
		AppVersion:    item.AppRevision.Version,
	}, nil
}

func RemoveAppFromDeploymentGroup(ctx context.Context, dgroupId, appId string) (bridgeId string, err error) {
	dGroupAppRevisions, err := GetDeploymentGroupBridgeAppRevision(ctx, dgroupId, appId)
	if err != nil {
		return "", err
	}
	if dGroupAppRevisions.DeploymentGroupBridgeAppRevisions == nil || dGroupAppRevisions.DeploymentGroupBridgeAppRevisions.Items == nil {
		return "", fmt.Errorf("app '%s' not found in Deployment Group '%s' not found", appId, dgroupId)
	}

	if len(dGroupAppRevisions.DeploymentGroupBridgeAppRevisions.Items) != 1 {
		return "", fmt.Errorf("found %d matching Deployment Group bridges for App '%s' in Deployment Group '%s', but expected 1", len(dGroupAppRevisions.DeploymentGroupBridgeAppRevisions.Items), appId, dgroupId)
	}

	id := dGroupAppRevisions.DeploymentGroupBridgeAppRevisions.Items[0].Id
	_, err = deleteDeploymentGroupBridgeAppRevision(ctx, id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func AddAppRevisionToDeploymentGroup(ctx context.Context, dgroupId, appId, appRevisionId string, modifyExisting bool) (bridgeId string, err error) {
	// process for existing bridges
	if res, _ := GetDeploymentGroupAppRevision(ctx, dgroupId, appId); res != nil {
		if res.AppRevisionId == appRevisionId {
			return res.Id, nil
		}
		if !modifyExisting {
			return "", fmt.Errorf("Deployment Group %s already contains app %s in revision %s", dgroupId, appId, appRevisionId)
		}
		if _, err = RemoveAppFromDeploymentGroup(ctx, dgroupId, appId); err != nil {
			return "", err
		}
	}

	// create new bridge
	res, err := AddAppRevisionToDeploymentGroupMutation(ctx, dgroupId, appRevisionId)
	if err != nil {
		return "", err
	}
	i := res.GetCreateDeploymentGroupBridgeAppRevisions()
	if len(i) != 1 {
		return "", fmt.Errorf("found %d matching Deployment Groups", len(i))
	}
	return i[0].Id, nil
}
