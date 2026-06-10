package asset

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:   "modify <id>",
	Short: "modify an Asset",
	Args:  cobra.ExactArgs(1),
	RunE:  runModifyCmd,
}

func init() {
	modifyCmd.Flags().StringP("name", "n", "", "name")
	modifyCmd.Flags().StringP("attestation-key", "k", "", "attestation key")
	AssetCmd.AddCommand(modifyCmd)
}

func runModifyCmd(cmd *cobra.Command, args []string) (err error) {
	id := args[0]

	asset, err := backend.GetAssetById(cmd.Context(), id)
	if err != nil {
		return err
	}
	if asset == nil {
		return fmt.Errorf("asset not found")
	}
	if tools.MaybeStringToString(asset.AssetType, "") != structs.AssetTypeSystem &&
		tools.MaybeStringToString(asset.AssetType, "") != structs.AssetTypeEd &&
		tools.MaybeStringToString(asset.AssetType, "") != structs.AssetTypeRtd {
		return fmt.Errorf("asset %s is not System", id)
	}

	var attestationKey *string
	if cmd.Flags().Changed("attestation-key") {
		if tools.MaybeStringToString(asset.AssetType, "") != structs.AssetTypeEd {
			return fmt.Errorf("only Assets of Edge Devices have an attestation key")
		}
		attestationKey = tools.StrPtr(cmd.Flag("attestation-key").Value.String())

	}

	err = backend.UpdateAsset(cmd.Context(), id, tools.MaybeFlagToMaybeString(cmd, "name"), attestationKey)
	asset, err = backend.GetAssetById(cmd.Context(), id)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, asset, infoFormatter)
}
