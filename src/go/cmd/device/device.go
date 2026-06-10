package device

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd"
)

var DeviceCmd = &cobra.Command{
	Use:     "device",
	Aliases: []string{"d"},
	Short:   "Manage devices",
}

func init() {
	cmd.RootCmd.AddCommand(DeviceCmd)
}
