package ssh

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync <deviceSerial or deviceName>",
	Short: "Sync the current user onto the device. This will remove the user, if he is already exists, and creates the user again. Only required when the user changed its SSH key",
	Args:  cobra.ExactArgs(1),
	RunE:  runSyncCmd,
}

func init() {
	sshCmd.AddCommand(syncCmd)
}

type deviceSshUserResetResult struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func runSyncCmd(cmd *cobra.Command, args []string) error {
	device, err := getDevice(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	reset, err := backend.DeviceSshUserReset(cmd.Context(), device.DeviceSerial)
	if err != nil {
		return err
	}

	msg := deviceSshUserResetResult{
		Message: reset.EdgeDeviceSshResetUser.Message,
		Success: reset.EdgeDeviceSshResetUser.Success,
	}
	return format.PrintFormattedOutput(cmd, msg, customSshUserResetFormatter)
}

func customSshUserResetFormatter(res deviceSshUserResetResult) (string, error) {
	if res.Success {
		return fmt.Sprintf("User reset successfully"), nil
	}
	return fmt.Sprintf("User reset failed: %s", res.Message), nil
}
