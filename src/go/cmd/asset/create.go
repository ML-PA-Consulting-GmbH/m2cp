package asset

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <model-id> <asset-name> [asset-description]",
	Short: "Create a new Edge Device Asset or Real Time Device Asset for a Device or System based on a given asset model",
	Args:  cobra.RangeArgs(2, 3),
	RunE:  runCreateCmd,
}

func init() {
	createCmd.Flags().String("attestation-key", "", "The attestation key")
	createCmd.Flags().String("hw-serial", "", "The hardware serial of the device asset, required for creating device assets")
	createCmd.Flags().String("mcu-id", "", "The hardware mcu-id")
	AssetCmd.AddCommand(createCmd)

}

func runCreateCmd(cmd *cobra.Command, args []string) (err error) {
	modelId := args[0]
	assetName := args[1]
	var assetDescription *string
	if len(args) > 2 {
		assetDescription = &args[2]
	}

	attestationKey := tools.FlagToMaybeString(cmd, "attestation-key")
	hwSerial := tools.MaybeStringToString(tools.FlagToMaybeString(cmd, "hw-serial"), "")
	mcuId := tools.FlagToMaybeString(cmd, "mcu-id")

	model, err := backend.GetAssetModelById(cmd.Context(), modelId)
	if err != nil {
		return err
	}
	if model.Type != structs.AssetTypeRtd && model.Type != structs.AssetTypeEd {
		return fmt.Errorf("asset type is invalid - this command is only for creating a new Asset for a Real Time Device or Edge Device")
	}

	id, err := backend.CreateAsset(cmd.Context(), modelId, hwSerial, mcuId, assetName, assetDescription, attestationKey)
	if err != nil {
		return err
	}

	asset, err := backend.GetAssetById(cmd.Context(), id)
	if err != nil {
		return err
	}
	if asset == nil {
		return fmt.Errorf("asset not found")
	}

	return format.PrintFormattedOutput(cmd, asset, infoFormatter)
}
