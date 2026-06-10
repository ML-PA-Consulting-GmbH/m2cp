package model

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"m2cpcli/tools/console"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <id>",
	Short: "info on a asset model",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfoCmd,
}

func init() {
	AssetModelCmd.AddCommand(infoCmd)
}

func runInfoCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("invalid number of arguments. Expected 1, got %d", len(args))
	}
	query := args[0]

	assetModel, err := backend.GetAssetModelById(cmd.Context(), query)
	if err != nil {
		return err
	}
	if assetModel == nil {
		return fmt.Errorf("asset model not found")
	}

	return format.PrintFormattedOutput(cmd, assetModel, infoFormatter)
}

func infoFormatter(assetModel *structs.AssetModel) (string, error) {
	out := ""
	out += fmt.Sprintf("\n%s:  ", console.Colorize(console.Green, "Asset Model"))
	out += fmt.Sprintf("\n  ├─ Asset Model Id:           %s", assetModel.Id)
	out += fmt.Sprintf("\n  ├─ Asset Model Name:         %s", assetModel.Name)
	out += fmt.Sprintf("\n  ├─ Asset Model Description:  %s", tools.MaybeStringToString(assetModel.Description, "n/a"))
	out += fmt.Sprintf("\n  └─ Tenant                    %s", assetModel.TenantName)

	out += "\n"

	return out, nil
}
