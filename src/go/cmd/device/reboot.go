package device

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var rebootCmd = &cobra.Command{
	Use:   "reboot <deviceId|deviceSerial|deviceName>",
	Short: "reboot a Device",
	Args:  cobra.ExactArgs(1),
	RunE:  runRebootCmd,
}

func init() {
	DeviceCmd.AddCommand(rebootCmd)
}

func runRebootCmd(cmd *cobra.Command, args []string) error {
	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	device, err := backend.GetDeviceInfoById(cmd.Context(), deviceId)
	if err != nil {
		return err
	}
	if device.IsRTD() {
		result, err := CoapRequest(cmd.Context(), device, "PUT", "/reboot/1", "", nil)
		if err != nil {
			return err
		}
		return format.PrintFormattedOutput(cmd, *result, nil)
	} else if device.IsED() {
		result, err := backend.DeviceRpc(cmd.Context(), "rpc.m2cp-gateway."+device.DeviceSerial, "DeviceReboot", nil)
		if err != nil {
			return err
		}
		return format.PrintFormattedOutput(cmd, *result, nil)
	} else {
		return fmt.Errorf("device type '%s' is not supported", device.DeviceModelRevision.Type)
	}
}
