package device

import (
	"context"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"m2cpcli/tools/console"
	"sort"
	"strings"

	"github.com/plgd-dev/kit/codec/json"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var infoCmd = &cobra.Command{
	Use:   "info <deviceSerial | deviceName | hw-serial>",
	Short: "Get information about a device",
	Long: `Get information about a device from the store.
	Usage:
	$ m2cp device info <os_serial | name | hw_serial --model the_model>
	Examples:
	$ m2cp device info 3f751536-3a2c-4b2c-b338-26ff720c4833
	$ m2cp device info dev-test
	$ m2cp device info 12345 --model virtual-device

	You may look up device models using "m2cp store model list".
	`,
	Args: cobra.ExactArgs(1),
	RunE: runInfoCmd,
}

func init() {
	DeviceCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolP("live", "l", false, "Get live information about the device")
	infoCmd.Flags().StringP("model", "m", "", "Device model (required when using hardware serial)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// TODO: flags
	//infoCmd.Flags().String("snaps", "", "list snaps currently installed, target installation and required actions")
	//infoCmd.Flags().String("users", "", "list users of device")
}

func IsHardwareSerial(s string) bool {
	if len(s) == 36 && strings.Count(s, "-") == 4 {
		return false
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}

	return len(s) > 0
}

const (
	actionRemove = "remove"
	actionAdd    = "sync"
	actionNone   = ""
	roleOwner    = "owner"
	roleCoAdmin  = "co-admin"
)

type SnapTableRow struct {
	Base        string   `json:"base"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Revision    int32    `json:"revision"`
	RevisionId  gql.UUID `json:"revision-id"`
	Description string   `json:"description"`
}

type User struct {
	//Id       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type DeviceInfo struct {
	IpAddresses    map[string][]string `json:"ipaddresses"`
	Users          []User              `json:"users"`
	HardwareSerial string              `json:"hardware-serial"`
	//ListedSnaps    interface{} `json:"listedsnaps"`
	//Result         string      `json:"result"`
	//Message        string      `json:"message"`
	//Nodes          interface{} `json:"nodes"`
}

type deviceInfoOutput struct {
	DeviceInfo        *structs.Device
	PendingActions    []structs.DevicePendingAction
	LiveInfo          *LiveInformation
	TotalDownloadSize int64
	System            *structs.Asset
}

func deviceInfoRpcCall(ctx context.Context, address string) (*DeviceInfo, error) {
	rpcInput := gql.ExecuteRpcInput{
		Command: "DeviceInfo",
		Address: address,
	}

	err := helper.ValidateRpcInput(&rpcInput)
	if err != nil {
		return nil, fmt.Errorf("invalid input: %s", err)
	}

	result, err := gql.ExecuteRpc(ctx, &rpcInput)
	if err != nil {
		return nil, fmt.Errorf("could not execute RPC: %s", err)
	}

	jsonBytes := []byte(result.Responses[0].Results[0].Value)

	di := DeviceInfo{}
	err = json.Decode(jsonBytes, &di)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON: %s", err)
	}

	return &di, nil
}

type LiveInformation struct {
	Warning        string              `json:"warning" yaml:"Warning,omitempty"`
	HardwareSerial string              `json:"hardware-serial" yaml:"Hardware-Serial,omitempty"`
	IpAddresses    map[string][]string `json:"ipaddresses" yaml:"IP-Addresses,omitempty"`
	Users          []User              `json:"users" yaml:"Users,omitempty"`
	RtdWhoami1     CoapResultType      `json:"rtd-whoami-1" yaml:"Rtd-Whoami-1"`
	RtdWhoami2     CoapResultType      `json:"rtd-whoami-2" yaml:"Rtd-Whoami-2"`
}

func runInfoCmd(cmd *cobra.Command, args []string) error {

	arg := args[0]
	modelFlag, _ := cmd.Flags().GetString("model")

	isOsSerial := func(s string) bool {
		return len(s) == 36 && strings.Count(s, "-") == 4
	}

	var deviceId string
	var err error

	switch {
	case isOsSerial(arg):

		deviceId, err = backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), arg)
		if err != nil {
			return err
		}

	case IsHardwareSerial(arg):

		if modelFlag == "" {
			return fmt.Errorf("when using a hardware serial, --model is required")
		}

		assets, err := backend.FindAsset(cmd.Context(), arg)
		if err != nil || len(assets) == 0 {
			return fmt.Errorf("no device found for hardware serial '%s' and model '%s'", arg, modelFlag)
		}

		for _, asset := range assets {
			if asset.Serial == arg &&
				asset.AssetModelName == modelFlag &&
				asset.DeviceId != nil {

				deviceId = *asset.DeviceId
				break
			}
		}

		if deviceId == "" {
			return fmt.Errorf("no device found for hardware serial '%s' and model '%s'", arg, modelFlag)
		}

	default:

		deviceId, err = backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), arg)
		if err != nil {
			return err
		}
	}

	deviceInfo, err := backend.GetDeviceInfoById(cmd.Context(), deviceId)
	if err != nil {
		return err
	}

	output := deviceInfoOutput{
		DeviceInfo:     deviceInfo,
		PendingActions: []structs.DevicePendingAction{},
	}

	if deviceInfo.DeviceAssetId != nil {
		output.System, err = backend.GetSystemByChildAssetId(cmd.Context(), *deviceInfo.DeviceAssetId)
		if err != nil {
			return err
		}
	}

	if deviceInfo.DevicePendingActions != nil &&
		len(deviceInfo.DevicePendingActions) > 0 {

		output.PendingActions =
			deviceInfo.DevicePendingActions
	}

	if cmd.Flag("live").Value.String() == "true" {
		output.LiveInfo =
			fetchEdgeDeviceLiveInfo(
				cmd.Context(),
				deviceInfo.DeviceSerial)
	}

	return format.PrintFormattedOutput(
		cmd,
		output,
		customDeviceInfoFormatter)
}

func fetchEdgeDeviceLiveInfo(ctx context.Context, deviceSerial string) *LiveInformation {
	di, err := deviceInfoRpcCall(ctx, fmt.Sprintf("rpc.m2cp-gateway.%s", deviceSerial))
	if err != nil {
		rpc := LiveInformation{
			Warning: err.Error(),
		}
		return &rpc
	}
	sort.Slice(di.Users, func(i, j int) bool {
		return di.Users[i].Email < di.Users[j].Email
	})

	rpc := LiveInformation{
		IpAddresses:    di.IpAddresses,
		HardwareSerial: di.HardwareSerial,
		Users:          di.Users,
	}
	return &rpc

}

func customDeviceInfoFormatter(output deviceInfoOutput) (string, error) {
	deviceInfo := output.DeviceInfo
	pendingActions := output.PendingActions
	system := output.System

	out := ""
	out += fmt.Sprintf("\n%s:  ", console.Colorize(console.Green, "Device"))
	out += fmt.Sprintf("\n  ├─ Id:                      %s", deviceInfo.DeviceId)
	out += fmt.Sprintf("\n  ├─ OS Serial:               %s", deviceInfo.DeviceSerial)
	out += fmt.Sprintf("\n  ├─ Hardware Serial:         %s", tools.MaybeStringToString(deviceInfo.HardwareSerialNumber, "n/a"))
	out += fmt.Sprintf("\n  ├─ Name:                    %s", tools.MaybeStringToString(deviceInfo.DeviceName, "n/a"))
	out += fmt.Sprintf("\n  ├─ Description:             %s", tools.MaybeStringToString(deviceInfo.Description, "n/a"))
	out += fmt.Sprintf("\n  ├─ Device Enabled:          %s", tools.BoolToString(deviceInfo.IsDeviceActivated, "yes", "no"))
	out += fmt.Sprintf("\n  ├─ Updates Enabled:         %s", tools.BoolToString(deviceInfo.IsUpdateActivated, "yes", "no"))
	out += fmt.Sprintf("\n  ├─ Last App Store Activity: %s", tools.MaybeTimeToString(deviceInfo.LastAppstoreActivity, "2006-01-02 15:04:05", "n/a"))
	out += fmt.Sprintf("\n  ├─ Last Messaging Activity: %s", tools.MaybeTimeToString(deviceInfo.LastMessagingActivity, "2006-01-02 15:04:05", "n/a"))
	out += fmt.Sprintf("\n  ├─ Last Messaging Hub:      %s", tools.MaybeStringToString(deviceInfo.LastHubEndpoint, "n/a"))
	out += fmt.Sprintf("\n  ├─ Last Messaging Mode:     %s", tools.MaybeStringToString(deviceInfo.UplinkMode, "n/a"))
	out += fmt.Sprintf("\n  ├─ Last Uptime:             %s", tools.MaybeInt64ToStringFormat(deviceInfo.LastUptime, "n/a", formatSeconds))
	out += fmt.Sprintf("\n  ├─ Type:                    %s", deviceInfo.DeviceModelRevision.Type)
	out += fmt.Sprintf("\n  ├─ Architecture:            %s", deviceInfo.DeviceModelRevision.Architecture)
	out += fmt.Sprintf("\n  ├─ Model Id:                %s", deviceInfo.DeviceModelRevision.Id)
	out += fmt.Sprintf("\n  ├─ Model Name:              %s", deviceInfo.DeviceModelRevision.Name)
	out += fmt.Sprintf("\n  ├─ Model Revision:          %d", deviceInfo.DeviceModelRevision.Revision)
	if deviceInfo.DeviceModelRevision.Type == structs.AssetTypeEd {
		out += fmt.Sprintf("\n  ├─ Attestation Key:         %s", tools.MaybeStringToString(deviceInfo.DeviceAttestationKey, "n/a"))
	}
	out += fmt.Sprintf("\n  └─ Asset Id:                %s", tools.MaybeStringToString(deviceInfo.DeviceAssetId, "n/a"))
	out += "\n\n"

	if deviceInfo.IsED() {
		out += fmt.Sprintf("%s:\n", console.Colorize(console.Green, "Gateway Stats"))
		if deviceInfo.LastUplinkSignalContent == nil {
			out += fmt.Sprintf("%s\n", console.Colorize(console.Yellow, "* no gateway stats available"))
		} else {
			details := deviceInfo.LastUplinkSignalContent
			out += fmt.Sprintf("  ├─ Heartbeat Counter:       %d\n", details.N)
			out += fmt.Sprintf("  ├─ Uptime Seconds:          %s\n", tools.MaybeIntToString(details.Gateway.UptimeSeconds, "n/a"))
			out += fmt.Sprintf("  ├─ Message Hub Reachable:   %v\n", tools.MaybeBoolToString(details.MessageHub.Up, "yes", "no", "n/a"))
			out += fmt.Sprintf("  ├─ System CPU Usage:        %s%%\n", tools.MaybeIntToString(details.Gateway.SystemCpuUsage, "n/a"))
			out += fmt.Sprintf("  ├─ FS Free MB:              %s\n", tools.MaybeIntToString(details.Gateway.FileSystemFreeMb, "n/a"))
			out += fmt.Sprintf("  ├─ Gateway Mem Usage MB:    %s\n", tools.MaybeIntToString(details.Gateway.MemGatewayMB, "n/a"))
			out += fmt.Sprintf("  ├─ Uploaded KB:             %s\n", tools.MaybeIntToString(details.Gateway.SentKB, "n/a"))
			out += fmt.Sprintf("  ├─ Buffer Count:            %s\n", tools.MaybeIntToString(details.Gateway.BufferCount, "n/a"))
			out += fmt.Sprintf("  ├─ Buffer Size KB:          %s\n", tools.MaybeIntToString(details.Gateway.BufferSizeKB, "n/a"))
			out += fmt.Sprintf("  ├─ Buffer Bad KB:           %s\n", tools.MaybeIntToString(details.Gateway.BufferBadKB, "n/a"))
			out += fmt.Sprintf("  └─ Buffer Age Seconds:      %s\n", tools.MaybeIntToString(details.Gateway.BufferAgeSeconds, "n/a"))
		}
		out += "\n"
	}

	// System info -----------------------------------------------------------------------

	out += fmt.Sprintf("%s:\n", console.Colorize(console.Green, "System"))
	if deviceInfo.DeviceAssetId == nil {
		out += fmt.Sprintf("%s\n", console.Colorize(console.Yellow, "* no Asset registered for this Device"))
	} else if deviceInfo.DeviceAssetAccess != nil && *deviceInfo.DeviceAssetAccess == false {
		out += fmt.Sprintf("%s\n", console.Colorize(console.Yellow, "* unauthorized to query details on the Device Asset "))
	} else if system == nil {
		out += fmt.Sprintf("%s\n", console.Colorize(console.Yellow, "* no System assigned"))
	} else {

		out += fmt.Sprintf("  ├─ Asset Id:             %s\n", system.Id)
		out += fmt.Sprintf("  ├─ Asset Serial:         %s\n", system.Serial)
		out += fmt.Sprintf("  ├─ Asset Name:           %s\n", system.Name)
		out += fmt.Sprintf("  └─ Asset Description:    %s\n", tools.MaybeStringToString(system.Description, "n/a"))
		out += "\n"
		for _, childAsset := range system.Components {
			out += fmt.Sprintf("%s:\n", console.Colorize(console.Green, fmt.Sprintf("System Component '%s'", childAsset.Name)))
			out += fmt.Sprintf("  ├─ Asset Id:             %s\n", childAsset.Id)
			out += fmt.Sprintf("  ├─ Asset HW Serial:         %s\n", childAsset.Serial)
			out += fmt.Sprintf("  ├─ Asset HW MCU-ID:         %s\n", tools.MaybeStringToString(childAsset.McuId, "n/a"))
			out += fmt.Sprintf("  ├─ Asset Name:           %s\n", childAsset.Name)
			out += fmt.Sprintf("  ├─ Asset Type:           %s\n", tools.MaybeStringToString(childAsset.AssetType, "n/a"))
			out += fmt.Sprintf("  ├─ Asset Model:          %s\n", childAsset.AssetModelName)
			var deviceSerial = "n/a"
			if childAsset.Device != nil {
				deviceSerial = childAsset.Device.DeviceSerial
			}
			out += fmt.Sprintf("  └─ Device OS Serial:     %s\n", deviceSerial)
			out += "\n"
		}
	}

	// Network Discovery

	if deviceInfo.DeviceArchitecture == "ARM64" || deviceInfo.DeviceArchitecture == "AMD64" {
		out += fmt.Sprintf("%s\n", console.Colorize(console.Green, "Real Time Devices Seen"))
		if deviceInfo.LastRealTimeDevices == nil || len(deviceInfo.LastRealTimeDevices) == 0 {
			out += fmt.Sprintf("%s\n", console.Colorize(console.Yellow, "* no Real Time Devices seen recently"))
		} else {
			peersTable := format.NewTable(map[string]string{
				"serial":   "Serial",
				"hwserial": "Hardware Serial",
				"name":     "Name",
				"arch":     "Arch",
				"model":    "Model",
			})
			for _, device := range deviceInfo.LastRealTimeDevices {
				peersTable.AddRow(map[string]string{
					"serial":   device.DeviceSerial,
					"hwserial": tools.MaybeStringToString(device.HardwareSerialNumber, "n/a"),
					"name":     tools.MaybeStringToString(device.DeviceName, "n/a"),
					"arch":     device.DeviceArchitecture,
					"model":    fmt.Sprintf("%s (%d)", device.DeviceModelRevision.Name, device.DeviceModelRevision.Revision),
				})
			}
			out += peersTable.StringSelect([]string{"serial", "hwserial", "name", "arch", "model"})
		}
	} else if deviceInfo.DeviceArchitecture == "ARM32" {
		if deviceInfo.LastEdgeDevice == nil {
			out += fmt.Sprintf("%s\n", console.Colorize(console.Yellow, "* no Edge Device seen recently"))
		} else {
			out += fmt.Sprintf("%s:\n", console.Colorize(console.Green, "Last Seen Edge Device"))
			out += fmt.Sprintf("  ├─ OS Serial:      %s\n", deviceInfo.LastEdgeDevice.DeviceSerial)
			out += fmt.Sprintf("  └─ Name:           %s\n", tools.MaybeStringToString(deviceInfo.LastEdgeDevice.DeviceName, "n/a"))
			out += "\n"
		}
	}
	out += "\n"

	// DGroup info -----------------------------------------------------------------------

	out += fmt.Sprintf("%s:\n", console.Colorize(console.Green, "Deployment Group"))
	if deviceInfo.DeploymentGroup == nil {
		out += fmt.Sprintf("%s\n", console.Colorize(console.Yellow, "* no Deployment Group assigned"))
	} else {
		out += fmt.Sprintf("  ├─ Id:             %s\n", deviceInfo.DeploymentGroup.Id)
		out += fmt.Sprintf("  ├─ Name:           %s\n", deviceInfo.DeploymentGroup.Name)
		out += fmt.Sprintf("  ├─ Owner:          %s\n", deviceInfo.DeploymentGroup.Owner.Name+" ("+deviceInfo.DeploymentGroup.Owner.Email+")")
		out += fmt.Sprintf("  └─ Description:    %s\n", tools.MaybeStringToString(deviceInfo.DeploymentGroup.Description, "n/a"))
	}
	out += "\n"

	// Installed Apps -----------------------------------------------------------------------

	out += fmt.Sprintf("%s: \n", console.Colorize(console.Green, "Installed Apps"))
	if deviceInfo.DeviceInstallStates == nil || len(deviceInfo.DeviceInstallStates) == 0 {
		out += fmt.Sprintf("%s\n", console.Colorize(console.Yellow, "* No installed Apps"))
	} else {

		appsTable := format.NewTable(map[string]string{
			"system":   "System",
			"name":     "Name",
			"version":  "Version",
			"revision": "Revision",
			"desc":     "Description",
		})

		for _, installedApp := range deviceInfo.DeviceInstallStates {
			core := "-"
			if deviceInfo.DeviceModelRevision != nil && deviceInfo.DeviceModelRevision.SystemApps != nil {
				for _, systemApp := range deviceInfo.DeviceModelRevision.SystemApps {
					if systemApp.Id == installedApp.AppId {
						core = "yes"
						break
					}
				}
			}
			appsTable.AddRow(map[string]string{
				"system":   core,
				"name":     installedApp.AppName,
				"version":  installedApp.AppVersion,
				"revision": fmt.Sprintf("%d", installedApp.AppRevision),
				"desc":     strings.TrimSpace(tools.ShortenRight(tools.ReplaceLinebreaks(installedApp.AppDescription), 120)),
			})
			appsTable = appsTable.Sorts([]string{"system", "name"})
		}
		out += appsTable.StringSelect([]string{"system", "name", "version", "revision", "desc"})
	}
	out += "\n"

	// Pending Actions -----------------------------------------------------------------------

	out += fmt.Sprintf("%s: \n", console.Colorize(console.Green, "Pending Actions"))
	if len(pendingActions) > 0 {

		table := format.NewTable(map[string]string{
			"action": "Action",
			"app":    "App",
			"from":   "Current",
			"to":     "Target",
			"size":   "Download",
			"delta":  "Delta",
		})

		for _, action := range pendingActions {
			table.AddRow(map[string]string{
				"action": action.PendingActionName,
				"app":    action.AppName,
				"from":   fmt.Sprintf("%s (%s)", tools.MaybeStringToString(action.InstalledVersion, "-"), tools.MaybeIntToString(action.InstalledRevision, "-")),
				"to":     fmt.Sprintf("%s (%s)", tools.MaybeStringToString(action.TargetVersion, "-"), tools.MaybeIntToString(action.TargetRevision, "-")),
				"size":   tools.MaybeInt64ToStringFormat(action.DownloadSize, "n/a", tools.FormatFileSize),
				"delta":  tools.MaybeInt64ToStringFormat(action.DeltaFileSize, "n/a", tools.FormatFileSize),
			})
		}

		out += table.Sorts([]string{"action", "asc", "app", "asc"}).StringSelect([]string{"action", "app", "from", "to", "size", "delta"})
	} else {
		out += fmt.Sprintf("%s", console.Colorize(console.Yellow, "* No pending actions"))
	}
	out += "\n\n"

	// Live Information -----------------------------------------------------------------------

	if output.LiveInfo != nil {

		out += fmt.Sprintf("%s: \n", console.Colorize(console.Green, "Users on Device (live)"))
		table := format.NewTable(map[string]string{
			"name":   "Name",
			"email":  "E-Mail",
			"role":   "Role",
			"action": "Pending Action",
		})

		for _, user := range output.LiveInfo.Users {
			role, action := roleAndActionOfUser(user.Email, deviceInfo.DeploymentGroup)
			table.AddRow(map[string]string{
				"name":   user.Username,
				"email":  user.Email,
				"role":   role,
				"action": action,
			})
		}
		if deviceInfo.DeploymentGroup != nil {
			for _, coAdmin := range deviceInfo.DeploymentGroup.CoOwners {
				action := actionOfAdmin(coAdmin.Email, output.LiveInfo.Users)
				if action != nil {
					table.AddRow(map[string]string{
						"name":   coAdmin.Name,
						"email":  coAdmin.Email,
						"role":   roleCoAdmin,
						"action": *action,
					})
				}
			}
		}
		out += table.Sorts([]string{"name", "email"}).StringSelect([]string{"role", "name", "email", "action"})
		out += "\n"

		out += fmt.Sprintf("%s:\n", console.Colorize(console.Green, "IP-Addresses (live)"))
		if output.LiveInfo.IpAddresses == nil {
			out += fmt.Sprintf("%s\n\n", console.Colorize(console.Yellow, "* No Ip Addresses Reported"))
		} else {
			yamlBytes, err := yaml.Marshal(output.LiveInfo.IpAddresses)
			if err != nil {
				return "", err
			}
			out += fmt.Sprintf("%s\n\n", string(yamlBytes))
		}
	}

	return out, nil
}

func actionOfAdmin(email string, users []User) *string {
	for _, user := range users {
		if user.Email == email {
			return nil
		}
	}
	return tools.StrPtr(actionAdd)
}

func roleAndActionOfUser(email string, dgroup *structs.DeploymentGroup) (role, action string) {
	const unauthorized = "unauthorized"
	if dgroup == nil {
		return unauthorized, actionRemove
	}
	if email == dgroup.Owner.Email {
		return roleOwner, actionNone
	}
	for _, coAdmin := range dgroup.CoOwners {
		if email == coAdmin.Email {
			return roleCoAdmin, actionNone
		}
	}
	return unauthorized, actionRemove
}

func formatSeconds(seconds int64) string {
	// Calculate days, hours, minutes, and seconds
	days := seconds / (24 * 3600)
	seconds = seconds % (24 * 3600)
	hours := seconds / 3600
	seconds = seconds % 3600
	minutes := seconds / 60
	seconds = seconds % 60

	result := ""
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm ", minutes)
	}
	if seconds > 0 {
		result += fmt.Sprintf("%ds ", seconds)
	}

	return result
}
