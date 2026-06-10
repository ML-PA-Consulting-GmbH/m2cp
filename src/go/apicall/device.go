package apicall

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"m2cp"
	"m2cp/messages"
	"m2cpcli/env"
	"m2cpcli/structs"
)

type TargetDevice struct {
	e     *env.RuntimeEnvironment
	state *env.PersistedState
	Info  structs.DeviceInfoLegacy
}

func NewTargetDevice(e *env.RuntimeEnvironment, state *env.PersistedState, serialOrName string, customHub string) (*TargetDevice, error) {
	device, err := GetDeviceInfo(e, state, serialOrName)
	if err != nil {
		return nil, err
	}
	if customHub != "" {
		device.DeviceInfo.Device.Hub = customHub
	}
	return &TargetDevice{
		e:     e,
		state: state,
		Info:  device.DeviceInfo,
	}, nil
}

func (d *TargetDevice) Rpc(msg m2cp.CommandMessage) (m2cp.ResponseMessage, *string, error) {
	var err error

	msgJson, err := messages.ToJson(msg)
	if err != nil {
		return nil, nil, err
	}

	type rpcRequest struct {
		Macaroon       string `json:"macaroon"`
		CommandMessage string `json:"commandMessage"`
	}
	type rpcResult struct {
		Result          string `json:"result"`
		Message         string `json:"message"`
		ResponseMessage string `json:"responseMessage"`
	}

	// send command message to hub, which will inject it into the m2cp messaging network
	postBody, err := json.Marshal(rpcRequest{
		CommandMessage: base64.StdEncoding.EncodeToString(msgJson),
		Macaroon:       "", //d.state.Macaroon,
	})
	if err != nil {
		return nil, nil, err
	}
	url := d.Info.Device.Hub + "/device/command"
	res, err := PostJson[rpcResult](d.e, url, postBody, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed POST device hub: %s", err)
	}
	if res.Result != "success" {
		return nil, nil, fmt.Errorf("failed calling device: %s", res.Message)
	}

	responseJson, err := base64.StdEncoding.DecodeString(res.ResponseMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("could not decode base64: %s", err)
	}

	responseJsonString := string(responseJson)
	responseMessage, err := messages.ResponseFromJson(responseJson)
	if err != nil {
		return nil, nil, fmt.Errorf("could transform JSON to object: %s", err)
	}
	return responseMessage, &responseJsonString, nil
}

func GetDeviceSerial(e *env.RuntimeEnvironment, state *env.PersistedState, device string) (string, error) {
	deviceInfo, err := GetDeviceInfo(e, state, device)
	if err != nil {
		return "", err
	}
	return deviceInfo.DeviceInfo.Device.Serial, nil
}

func GetDeviceInfo(e *env.RuntimeEnvironment, state *env.PersistedState, device string) (*structs.PostDeviceInfoResponse, error) {
	req, err := json.Marshal(structs.PostDeviceInfoRequest{
		Device: device,
	})
	if err != nil {
		return nil, fmt.Errorf("failed creating request")
	}
	result, err := PostJson[structs.PostDeviceInfoResponse](e, MlpaDeviceInfoCmd, req, state.GetAuthHeader())
	if err != nil {
		return nil, fmt.Errorf("device '%s' not found in store (error: %s)", device, err.Error())
	}
	if result.Result != "success" {
		return nil, fmt.Errorf("failed getting device info for device '%s', taskId='%s'", device, result.Message)
	}
	return result, nil
}

func GetLiveDeviceInfo(e *env.RuntimeEnvironment, state *env.PersistedState, hub, serial string) (*structs.DeviceInfoResult, error) {

	var err error
	// get live device info
	postBody, err := json.Marshal(structs.DeviceInfoRequest{
		DeviceSerial: serial,
		Macaroon:     "", //state.Macaroon,
	})
	if err != nil {
		return nil, err
	}
	type deviceInfoEncoded struct {
		Result     string `json:"result"`
		Message    string `json:"message"`
		DeviceInfo string `json:"deviceInfo"`
	}
	apiUrl := hub + "/device/info"
	res, err := PostJson[deviceInfoEncoded](e, apiUrl, postBody, nil)
	if err != nil {
		fmt.Printf("ERROR     : %s\n", err.Error())
		fmt.Printf("url       : %s\n", apiUrl)
		fmt.Printf("post-body : %s\n", postBody)
		return nil, err
	}
	if res.Result != "success" {
		return nil, fmt.Errorf("can't reach device for live status: %s\n", res.Message)
	}

	// result is a base64 encoded json - let's unpack
	jsonRaw, err := base64.StdEncoding.DecodeString(res.DeviceInfo)
	if err != nil {
		return nil, fmt.Errorf("failed decoding base64 result from device hub")
	}
	var deviceInfo structs.DeviceInfoResult
	err = json.Unmarshal(jsonRaw, &deviceInfo)
	return &deviceInfo, err
}

func GetDeviceRefresh(e *env.RuntimeEnvironment, state *env.PersistedState, hub, deviceSerial string) (*structs.DeviceRefreshResult, error) {
	postBody, err := json.Marshal(structs.DeviceRefreshRequest{
		DeviceSerial: deviceSerial,
		Macaroon:     "", //state.Macaroon,
	})
	if err != nil {
		return nil, err
	}

	url := hub + "/device/refresh"
	res, err := PostJson[structs.DeviceRefreshResult](e, url, postBody, nil)
	return res, err
}

func ModifyDevice(e *env.RuntimeEnvironment, state *env.PersistedState, deviceSerial, name, description string, updateEnabled bool) error {
	var fields []string
	// We use a "?" as `nil` value. The empty string is allowed to be sent, to unset a virtual device name prior deletion.
	if name != "?" {
		fields = append(fields, "name")
	}
	if description != "?" {
		fields = append(fields, "description")
	}
	fmt.Println(fmt.Sprintf("WARNING: updateEnabled=%t is not implemented yet!", updateEnabled)) // TODO: implement on server side!

	reqObj := structs.PostMlpaDeviceModifyRequest{
		Serial:      deviceSerial,
		Name:        name,
		Description: description,
		Fields:      fields,
	}
	req, err := json.Marshal(reqObj)
	if err != nil {
		return fmt.Errorf("on marshalling the request: %s", err)
	}
	response, err := PostJson[structs.PostMlpaDeviceModifyResponse](e, MlpaDeviceModifyCmd, req, state.GetAuthHeader())
	if err != nil {
		return fmt.Errorf("on receiving the response: %s", err)
	} else {
		if response.Result != "success" {
			return fmt.Errorf("request ended with: %s, %s", response.Result, response.Message)
		}
	}
	return nil
}
