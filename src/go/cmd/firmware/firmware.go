package firmware

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd"
)

var FirmwareCmd = &cobra.Command{
	Use:     "firmware",
	Aliases: []string{"f"},
	Short:   "Manage firmware",
}

func init() {
	cmd.RootCmd.AddCommand(FirmwareCmd)
}
