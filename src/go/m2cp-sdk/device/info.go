package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"m2cp"
	coap_client "m2cp/coap/coap-client"
	"m2cp/snapd"
	"m2cp/tools"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

/**

TODO: This is a list of device properties, that should be available through the SDK, as they're reporting is mandatory for the M2CP platform.

Device Name
Device Type
Device Model Type
Device Serial Number
Device Part Number
Device Hardware Revision
Device Build Number
Device Bootloader Info
Device OS Version
Device Firmware Version
Device Active Firmware
Device Application Version
Device Factory Data
Device Custom Data
Device Test Data
Device Configuration Data

*/

func GetName(ctp m2cp.ContextPlus) (string, error) {
	var err error

	// If `M2CP_VIRTUAL_DEVICE` is set, we have to use the specified port to get the serial.
	// This is the case, if the application should be debugged in a virtual device container from the users host system
	m2cpVirtualDeviceDebugPort, isVirtualDeviceDebugging := os.LookupEnv("M2CP_VIRTUAL_DEVICE")
	if isVirtualDeviceDebugging {
		var baseSnapdPort int
		if baseSnapdPort, err = strconv.Atoi(m2cpVirtualDeviceDebugPort); err != nil {
			panic(fmt.Sprintf("failed parsing M2CP_VIRTUAL_DEVICE: %s", err))
		}
		if serial, err := getSnapdSerialFromUrl(ctp, baseSnapdPort); err != nil {
			return "", fmt.Errorf("failed getting snapd serial from virtual device port %d: %s", baseSnapdPort, err)
		} else {
			ctp.LogDebug("got serial from snapd in virtual device: %s", serial)
			return serial, nil
		}
	}

	// Get serial from snapCtl, which is available and accessible in all executed snaps
	// https://snapcraft.io/docs/using-snapctl
	serial, err := getSnapdSerialFromSnapCtl()
	if err == nil {
		ctp.LogDebug("got serial from snapctl: %s", serial)
		return serial, nil
	}
	ctp.LogDebug("failed getting serial from snapctl: %s", err)

	// If hostname is a valid UUID, this is our serial
	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	if _, err := uuid.Parse(hostName); err == nil {
		ctp.LogDebug("got serial from hostname: %s", hostName)
		return hostName, nil
	}
	ctp.LogDebug("failed getting serial from hostname, not a valid UUID: %s", hostName)

	// Last chance: if we're in a CI, we'll mock our uuid
	if tools.IsRunningInAzurePipeline() {
		ctp.LogDebug("running in azure pipeline - using mock uuid '00000000-0000-0000-0000-000000000000'")
		return "00000000-0000-0000-0000-000000000000", nil
	}

	return "", fmt.Errorf("could not get device serial from any known source")
}

func GetType() (string, error) {
	return "m2cp-device", nil
}

type ModelInfo struct {
	Model        string `json:"model"`
	BrandId      string `json:"brand_id"`
	Architecture string `json:"architecture"`
	Revision     int    `json:"revision"`
}

func GetModelInfo(ctp m2cp.ContextPlus) (ModelInfo, error) {
	assertion, err := snapd.GetModelAssertion(ctp)
	if err != nil {
		return ModelInfo{}, err
	}

	if assertion.Architecture == "" {
		assertion.Architecture = "amd64"
	}

	return ModelInfo{
		Model:        assertion.Model,
		BrandId:      assertion.BrandId,
		Architecture: assertion.Architecture,
		Revision:     assertion.Revision,
	}, nil
}

// GetModel get identifier encoding the model of the device.
func GetModel(ctp m2cp.ContextPlus) (string, error) {
	assertion, err := GetModelInfo(ctp)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s/%d", assertion.BrandId, assertion.Model, assertion.Architecture, assertion.Revision), nil
}

type coapDeviceInfo struct {
	Sn     uint32 `json:"sn"`
	Board  string `json:"board"`
	App    string `json:"app"`
	AppVer string `json:"app_ver"`
	AppRev int    `json:"app_rev"`
}

func GetHwSerial(ctp m2cp.ContextPlus) (string, error) {
	const errorPrefix = "failed hw serial request to m2cp-coap: %s"
	client, err := coap_client.NewClient(ctp, "::1")
	if err != nil {
		return "", fmt.Errorf(errorPrefix, err.Error())
	}
	res, err := client.Get("/whoami/1", "", nil)
	if err != nil {
		return "", fmt.Errorf(errorPrefix, err.Error())
	}

	payload := res.GetBody()

	var info coapDeviceInfo
	if err = json.Unmarshal(payload, &info); err != nil {
		return "", fmt.Errorf(errorPrefix, err.Error())
	}

	return tools.EncodeHardwareSerial(info.Sn), nil
}

// Deprecated: use tools.DecodeHardwareSerial instead.
//
//go:fix inline
func DecodeHwSerial(encodedID string) (uint32, error) {
	return tools.DecodeHardwareSerial(encodedID)
}

// Deprecated: use tools.EncodeHardwareSerial instead.
//
//go:fix inline
func EncodeHwSerial(id uint32) string {
	return tools.EncodeHardwareSerial(id)
}

func GetOsVersion(ctp m2cp.ContextPlus) (string, error) {
	return getOsVersion2(ctp)
}

func MakeOsVersion(ctp m2cp.ContextPlus, snaps map[string]uint, modelAssertion []byte) (string, error) {
	model, err := snapd.ParseModelAssertion(modelAssertion)
	if err != nil {
		return "", fmt.Errorf("could not parse model assertion: %s", err)
	}
	return calculateOsVersion2(ctp, snaps, *model)
}

func DecodeOsVersion(version string) (*OsVersion, error) {
	errorTemplate := "invalid os-version string: "

	tokens := strings.Split(version, "-")
	if len(tokens) < 2 {
		return nil, errors.New(errorTemplate + "too few tokens")
	}
	// first token specifies the encoding format revision
	format, err := strconv.Atoi(tokens[0])
	if err != nil {
		return nil, fmt.Errorf(errorTemplate+"unknown format: %s", tokens[0])
	}
	switch format {
	case 2:
		return decodeOsVersion2(version)
	}
	return nil, fmt.Errorf(errorTemplate+"unkown format revision: '%d'", format)
}
