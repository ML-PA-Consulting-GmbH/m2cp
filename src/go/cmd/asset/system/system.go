package system

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/asset"
)

var AssetSystemCmd = &cobra.Command{
	Use:     "system",
	Aliases: []string{"s"},
	Short:   "Manage system assets",
}

func init() {
	asset.AssetCmd.AddCommand(AssetSystemCmd)
}
