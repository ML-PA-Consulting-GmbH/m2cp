package node

import (
	"m2cpcli/cmd/device"

	"github.com/spf13/cobra"
)

var snapCmd = &cobra.Command{
	Use:   "snap",
	Short: "Remote snap commands",
}

func init() {
	device.DeviceCmd.AddCommand(snapCmd)
}
