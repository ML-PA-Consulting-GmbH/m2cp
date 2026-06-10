package asset

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd"
)

var AssetCmd = &cobra.Command{
	Use:     "asset",
	Aliases: []string{"a"},
	Short:   "Manage assets",
}

func init() {
	cmd.RootCmd.AddCommand(AssetCmd)
}
