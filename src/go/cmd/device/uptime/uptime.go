package uptime

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/device"
)

var uptimeCmd = &cobra.Command{
	Use:     "uptime",
	Aliases: []string{"n"},
	Short:   "Work with device uptime",
}

func init() {
	device.DeviceCmd.AddCommand(uptimeCmd)
}
