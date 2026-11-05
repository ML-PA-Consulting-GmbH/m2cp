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

// TODO: fix duplicate code, only differs by signature type
func IsRelevantAssetType(asset *structs.Asset) bool {
	switch asset.AssetTypeId {
	case structs.AssetModelTypeIdSystem, structs.AssetModelTypeIdEd, structs.AssetModelTypeIdRtd:
		return true
	default:
		return false
	}
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

	if !IsRelevantAssetType(asset) {
		return fmt.Errorf("wrong asset type id %s", asset.AssetTypeId)
	}

	var attestationKey *string
	if cmd.Flags().Changed("attestation-key") {
		if asset.AssetTypeId != structs.AssetModelTypeIdEd {
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
