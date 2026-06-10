package model

import (
	"m2cpcli/cmd/asset"

	"github.com/spf13/cobra"
)

var AssetModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage asset models",
}

func init() {
	asset.AssetCmd.AddCommand(AssetModelCmd)
}
