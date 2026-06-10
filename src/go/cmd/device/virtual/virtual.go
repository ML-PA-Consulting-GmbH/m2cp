package virtual

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/device"
)

var VirtualDeviceCmd = &cobra.Command{
	Use:     "virtual",
	Aliases: []string{"v"},
	Short:   "Manage virtual devices",
}

func init() {
	device.DeviceCmd.AddCommand(VirtualDeviceCmd)
}
