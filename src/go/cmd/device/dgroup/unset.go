package dgroup

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var unsetCmd = &cobra.Command{
	Use:   "unset [deviceSerial|deviceName]",
	Short: "Remove the Device from a Deployment Group",
	Args:  cobra.ExactArgs(1),
	RunE:  runUnsetCmd,
}

func init() {
	fleetCmd.AddCommand(unsetCmd)
	dgroupCmd.AddCommand(unsetCmd)
}

type DGroupUnsetResult struct {
	Message        string      `json:"message"`
	ModifiedDevice interface{} `json:"modifiedDevice"`
	TaskId         string      `json:"task-id"`
}

func runUnsetCmd(cmd *cobra.Command, args []string) (err error) {
	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	res, err := backend.UnsetDeviceDeploymentGroup(cmd.Context(), deviceId)
	if err != nil {
		return err
	}

	deviceHandle := res.UpdateDevices[0].SerialNumber
	if res.UpdateDevices[0].DeviceName != nil {
		deviceHandle += " (" + *res.UpdateDevices[0].DeviceName + ")"
	}

	msg := DGroupUnsetResult{
		Message:        fmt.Sprintf("removed Device %s from any Deployment Group", deviceHandle),
		ModifiedDevice: res.UpdateDevices[0],
	}
	return format.PrintFormattedOutput(cmd, msg, customDGroupUnsetFormatter)
}

func customDGroupUnsetFormatter(res DGroupUnsetResult) (string, error) {
	return res.Message, nil
}
