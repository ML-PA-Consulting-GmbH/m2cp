package dgroup

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [deviceSerial|deviceName dgroupName|dgroupId]",
	Short: "set the Deployment Group of a Device",
	Args:  cobra.ExactArgs(2),
	RunE:  runSetCmd,
}

type DGroupSetResult struct {
	Message        string      `json:"message"`
	ModifiedDevice interface{} `json:"modifiedDevice"`
	TaskId         string      `json:"task-id"`
}

func init() {
	fleetCmd.AddCommand(setCmd)
	dgroupCmd.AddCommand(setCmd)
}

func runSetCmd(cmd *cobra.Command, args []string) (err error) {
	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	dgroup, err := backend.GetDeploymentGroupByNameOrId(cmd.Context(), args[1])
	if err != nil {
		return err
	}

	res, err := backend.SetDeviceDeploymentGroup(cmd.Context(), deviceId, &dgroup.Id)
	if err != nil {
		return err
	}

	deviceHandle := res.UpdateDevices[0].SerialNumber
	if res.UpdateDevices[0].DeviceName != nil {
		deviceHandle += " (" + *res.UpdateDevices[0].DeviceName + ")"
	}
	dgroupHandle := *res.UpdateDevices[0].DeploymentGroupId
	if res.UpdateDevices[0].DeploymentGroup != nil {
		dgroupHandle += " (" + res.UpdateDevices[0].DeploymentGroup.Name + ")"
	}
	msg := DGroupSetResult{
		Message:        fmt.Sprintf("assigned Device %s to Deployment Group %s", deviceHandle, dgroupHandle),
		ModifiedDevice: res.UpdateDevices,
	}

	return format.PrintFormattedOutput(cmd, msg, customFleetSetFormatter)
}

func customFleetSetFormatter(res DGroupSetResult) (string, error) {
	return res.Message, nil
}
