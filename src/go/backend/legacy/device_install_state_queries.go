package legacy

import (
	"context"
	"fmt"
)

type DeviceInstallState struct {
	Id                string
	EdgeDeviceId      string
	SnapRevisionId    string
	SnapDeclarationId string
}

func GetDevicesInstallStates(ctx context.Context, deviceIds []string) (map[string][]DeviceInstallState, error) {
	if deviceIds == nil || len(deviceIds) == 0 {
		return nil, fmt.Errorf("deviceIds is required")
	}

	take := 1000
	skip := 0

	result := make(map[string][]DeviceInstallState)

	for {
		response, err := getDeviceInstallStateByDeviceIds(ctx, take, skip, deviceIds)
		if err != nil {
			return nil, err
		}

		for _, item := range response.Data.Items {
			deviceId := item.EdgeDeviceId
			result[deviceId] = append(result[deviceId], DeviceInstallState{
				Id:                item.Id,
				EdgeDeviceId:      item.EdgeDeviceId,
				SnapRevisionId:    item.SnapRevisionId,
				SnapDeclarationId: item.SnapRevision.SnapDeclarationId,
			})
		}

		if !response.Data.PageInfo.HasNextPage {
			break
		}

		skip += take
	}

	return result, nil
}

func EdgeDeviceRefresh(ctx context.Context, osSerial string) error {
	res, err := edgeDeviceRefresh(ctx, osSerial)
	if err != nil {
		return err
	}
	_ = res
	return nil
}
