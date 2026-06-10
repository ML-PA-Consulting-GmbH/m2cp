package graphql

import (
	"context"
	"fmt"
	"m2cpcli/tools"
)

const DeviceModifyFieldName = "deviceName"
const DeviceModifyFieldDescription = "deviceDescription"

// DeviceModify modifies the name and/or description of a device. If the name or description is set to an empty string, the field will not be modified.
// To unset a field, add the field name to the setToNull slice. Available fields are DeviceModifyFieldName and DeviceModifyFieldDescription. If no field should be unset, setToNull can be nil.
func DeviceModify(ctx context.Context, deviceId UUID, updateName string, updateDescription string, setToNull []string) (*EdgeDeviceModifySubset, error) {
	mutationTemplate := `mutation($id: UUID!){
		updateEdgeDevices(edgeDevices: [{
			id: $id
			{{ if .Name }}
			deviceName: {{ printf "%q" .Name }}
			{{ end }}
			{{ if .Description }}
			deviceDescription: {{ printf "%q" .Description }}
			{{ end }}
		}],
		{{ if .SetNullFields }}
		setNull: {
			{{ range .SetNullFields }}
			{{ . }}: true,
			{{ end }}
		}
		{{ end }}
		) {
			deviceSerial
			deviceName
			deviceDescription
			fleetId
		}
	}`
	type Parameters struct {
		Name          string
		Description   string
		SetNullFields []string
	}
	params := Parameters{}

	if updateName != "" {
		params.Name = updateName
	}
	if updateDescription != "" {
		params.Description = updateDescription
	}

	if setToNull != nil {
		for _, field := range setToNull {
			params.SetNullFields = append(params.SetNullFields, field)
		}
	}

	mutationString, err := RenderQueryOrMutation(mutationTemplate, params)
	if err != nil {
		return nil, err
	}

	client, req := PrepareClientAndRequest(ctx, mutationString)
	req.Var("id", deviceId)

	var result struct {
		DeviceModify []EdgeDeviceModifySubset `json:"updateEdgeDevices"`
	}
	err = client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	if len(result.DeviceModify) != 1 {
		return nil, fmt.Errorf("unexpected number of devices modified: %d", len(result.DeviceModify))
	}
	return &result.DeviceModify[0], nil
}

func DeviceIdByNameOrSerial(ctx context.Context, deviceNameOrSerial string) (UUID, error) {
	var err error
	var deviceId UUID
	if tools.IsValidUuid(deviceNameOrSerial) {
		serial := deviceNameOrSerial
		deviceId, err = DeviceIdBySerial(ctx, serial)
		if err != nil {
			return deviceId, err
		}
	} else {
		name := deviceNameOrSerial
		deviceId, err = DeviceIdByName(ctx, name)
		if err != nil {
			return deviceId, err
		}
	}
	return deviceId, nil
}

func DeviceSerialByName(ctx context.Context, deviceName string) (string, error) {
	queryString := `query getEdgeDeviceSerial($name: String){
result: edgeDevices(where: {deviceName: {eq: $name}}, order: {id:ASC}){
    items{
		deviceSerial
    } 
    totalCount
}}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("name", deviceName)

	var result struct {
		EdgeDevices EdgeDeviceCollectionSegment `json:"result"`
	}

	err := client.Run(ctx, req, &result)

	if err != nil {
		return "", err
	}
	if result.EdgeDevices.TotalCount > 1 {
		return "", fmt.Errorf("deviceName \"%s\" is not unique", deviceName)
	}
	if result.EdgeDevices.TotalCount == 0 {
		return "", fmt.Errorf("deviceName \"%s\" not found", deviceName)
	}
	return result.EdgeDevices.Items[0].DeviceSerial, nil

}

func DeviceIdBySerial(ctx context.Context, deviceSerial string) (UUID, error) {
	queryString := `query($serial: String){
result: edgeDevices(where: {deviceSerial: {eq: $serial}}, order: {id:ASC}){
    items{
		id
    } 
    totalCount
}}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("serial", deviceSerial)

	var result struct {
		EdgeDevices EdgeDeviceCollectionSegment `json:"result"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return "", err
	}

	if result.EdgeDevices.TotalCount > 1 {
		return "", fmt.Errorf("deviceSerial \"%s\" is not unique", deviceSerial)
	}
	if result.EdgeDevices.TotalCount == 0 {
		return "", fmt.Errorf("deviceSerial \"%s\" not found", deviceSerial)
	}
	return result.EdgeDevices.Items[0].Id, nil
}

func DeviceIdByName(ctx context.Context, deviceName string) (UUID, error) {
	queryString := `query($name: String){
edgeDevices(where: {deviceName: {eq: $name}}, order: {id:ASC}){
    items{
		id
    } 
    totalCount
}}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("name", deviceName)

	var result struct {
		EdgeDevices EdgeDeviceCollectionSegment `json:"edgeDevices"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return "", err
	}

	if result.EdgeDevices.TotalCount > 1 {
		return "", fmt.Errorf("deviceName \"%s\" is not unique", deviceName)
	}
	if result.EdgeDevices.TotalCount == 0 {
		return "", fmt.Errorf("deviceName \"%s\" not found", deviceName)
	}
	return result.EdgeDevices.Items[0].Id, nil
}

func DeviceByDeviceId(ctx context.Context, deviceId UUID) (*EdgeDevice, error) {
	var err error

	queryString := `query($edgeDeviceId: UUID!){
	result: edgeDevice(id: $edgeDeviceId){
		id
	    deviceSerial
		deviceArchitecture
		deviceName
		deviceDescription
		deviceLastActivity
		isDeviceActivated
		fleetId
		fleet {
			id
			fleetName
			description
			architecture
    }}}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("edgeDeviceId", deviceId)

	var result struct {
		Device EdgeDevice `json:"result"`
	}

	err = client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return &result.Device, err
}

func EdgeDevicePing(ctx context.Context, deviceId UUID) (*EdgeDevicePingOutput, error) {
	device, err := DeviceByDeviceId(ctx, deviceId)
	if err != nil {
		return nil, err
	}

	queryString := `query($deviceSerial: String!) {
	result: edgeDevicePing(input: {deviceSerial: $deviceSerial}) {
		uptimeSeconds
	}}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("deviceSerial", device.DeviceSerial)

	var result struct {
		Result EdgeDevicePingOutput `json:"result"`
	}

	err = client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

func EdgeDeviceRefresh(ctx context.Context, deviceId UUID) (*EdgeDeviceRefreshOutput, error) {
	device, err := DeviceByDeviceId(ctx, deviceId)
	if err != nil {
		return nil, err
	}

	queryString := `mutation($deviceSerial: String!) {
	result: edgeDeviceRefresh(input: {deviceSerial: $deviceSerial}) {
		success
		message
	}}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("deviceSerial", device.DeviceSerial)

	var result struct {
		Result EdgeDeviceRefreshOutput `json:"result"`
	}

	err = client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

func EdgeDeviceInstallationStates(ctx context.Context, deviceId UUID) ([]EdgeDeviceInstallationState, error) {
	var result []EdgeDeviceInstallationState
	var err error

	queryString := `query GetInstallationStates($id: UUID!, $take: Int, $skip: Int) {
  result: edgeDeviceInstallStates(
    skip: $skip, take: $take, order: {id:ASC},
    where: {edgeDeviceId: {eq: $id}}
  ) {
    items{
      edgeDevice {
	    id
	    deviceSerial
	    deviceArchitecture
	    deviceName
	    deviceDescription
	    deviceLastActivity
	    isDeviceActivated
	    fleetId
	    fleet {
          id
	      fleetName
	      description
	      architecture
        }
      }
      edgeDeviceId
      id
      snapRevision {
        id
	    uploadMessage
	    revision
	    snapDownloadSize
	    createdAt
	    snapVersion
		snapDeclarationId
		snapDeclaration {
			snapDescription
		}
      }
      snapRevisionId
    } 
    totalCount
  }
}`

	client, request := PrepareClientAndRequest(ctx, queryString)
	request.Var("id", deviceId)

	result, err = RunPaginated[EdgeDeviceInstallationState](ctx, client, request)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func EdgeDeviceByDeviceId(ctx context.Context, deviceId UUID) (*EdgeDevice, error) {
	var err error

	queryString := `query($id: UUID!){
result: edgeDevice(id: $id){
	id
	deviceSerial
	deviceArchitecture
	deviceName
	deviceDescription
	deviceLastActivity
	isDeviceActivated
	fleetId
	fleet {
      id
	  fleetName
	  description
	  architecture
    }
	edgeDeviceModel {
      id
      modelArchitecture
      modelName
      modelRevision
      modelType      
      isTpmRequired
    }
}}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("id", deviceId)

	var queryResult struct {
		Device EdgeDevice `json:"result"`
	}
	err = client.Run(ctx, req, &queryResult)
	if err != nil {
		return nil, err
	}

	return &queryResult.Device, nil
}

func EdgeDeviceSshOpen(ctx context.Context, deviceSerial string, resetConnection bool) (*EdgeDeviceSshOpenOutput, error) {
	mutationString := `mutation sshOpen($deviceSerial: String!, $forceResetConnection: Boolean){
  edgeDeviceSshOpenV2(input: {
    deviceSerial: $deviceSerial
	forceResetConnection: $forceResetConnection
  })
  {
    host,
	remoteSshPort,
    reverseSshPort,
    usernameDevice,
	usernameServer
  }
}`
	client, req := PrepareClientAndRequest(ctx, mutationString)
	req.Var("deviceSerial", deviceSerial)
	req.Var("forceResetConnection", resetConnection)

	var result struct {
		Result EdgeDeviceSshOpenOutput `json:"edgeDeviceSshOpenV2"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}
	return &result.Result, nil
}

func EdgeDeviceSshClose(ctx context.Context, deviceSerial string, port string) (string, error) {
	mutationString :=
		`mutation sshClose($deviceSerial: String!, $port: String!){
  edgeDeviceSshClose(input: {
    deviceSerial: $deviceSerial,
    port: $port
  })
  {
    message
    port,
  }
}`

	client, req := PrepareClientAndRequest(ctx, mutationString)
	req.Var("deviceSerial", deviceSerial)
	req.Var("port", port)

	var result struct {
		Result EdgeDeviceSshCloseOutput `json:"edgeDeviceSshClose"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return "", err
	}
	return result.Result.Message, nil
}

func EdgeDevicesOnlineStatusByDeviceSerial(ctx context.Context, devices []EdgeDevice) (*[]EdgeDeviceOnlineStatus, error) {
	queryString := `query EdgeDeviceOnlineStatus($deviceSerials: [String!]!) {
  edgeDeviceOnlineStatus(deviceSerials: $deviceSerials) {
	deviceSerial
    connectionState
    isDeviceRegistered
    lastActivityTime
  }
}`

	var deviceSerials []string
	for _, device := range devices {
		deviceSerials = append(deviceSerials, device.DeviceSerial)
	}

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("deviceSerials", deviceSerials)

	var queryResult struct {
		EdgeDeviceOnlineStatus []EdgeDeviceOnlineStatus `json:"edgeDeviceOnlineStatus"`
	}

	err := client.Run(ctx, req, &queryResult)
	if err != nil {
		return nil, err
	}

	return &queryResult.EdgeDeviceOnlineStatus, nil

}
