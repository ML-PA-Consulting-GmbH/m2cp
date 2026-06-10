package node

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/device"
)

var nodeCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"n"},
	Short:   "Manage nodes",
}

func init() {
	device.DeviceCmd.AddCommand(nodeCmd)
}
