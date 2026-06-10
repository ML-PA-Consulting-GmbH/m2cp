package model

import (
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:   "modify <id>",
	Short: "modify an asset model",
	Args:  cobra.ExactArgs(1),
	RunE:  runModifyCmd,
}

func init() {
	modifyCmd.Flags().StringP("name", "n", "", "name")
	modifyCmd.Flags().StringP("description", "d", "", "description")
	AssetModelCmd.AddCommand(modifyCmd)
}

func runModifyCmd(cmd *cobra.Command, args []string) (err error) {
	id := args[0]

	err = backend.UpdateAssetModel(cmd.Context(),
		id,
		tools.MaybeFlagToMaybeString(cmd, "name"),
		tools.MaybeFlagToMaybeString(cmd, "description"),
	)
	if err != nil {
		return err
	}
	assetModel, err := backend.GetAssetModelById(cmd.Context(), id)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, assetModel, infoFormatter)
}
