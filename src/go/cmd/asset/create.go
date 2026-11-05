package asset

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <model-id> <asset-name> [asset-description]",
	Short: "Create a new Asset for the given Asset Model ID. Only assets of type Edge Device, Real Time Device, or System can be created",
	Args:  cobra.RangeArgs(2, 3),
	RunE:  runCreateCmd,
}

func init() {
	createCmd.Flags().String("attestation-key", "", "The attestation key, required for Edge Device assets")
	createCmd.Flags().String("hw-serial", "", "The hardware serial of the asset, required for creating Edge Device and Real Time Device assets")
	createCmd.Flags().String("mcu-id", "", "The hardware mcu-id, required for Real Time Device assets")
	AssetCmd.AddCommand(createCmd)

}

// TODO: fix duplicate code, only differs by signature type
func IsRelevantAssetModelType(assetModel *structs.AssetModel) bool {
	switch assetModel.TypeId {
	case structs.AssetModelTypeIdSystem, structs.AssetModelTypeIdEd, structs.AssetModelTypeIdRtd:
		return true
	default:
		return false
	}
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
	if !IsRelevantAssetModelType(model) {
		return fmt.Errorf("asset type %q is invalid - this command is only for creating a new Asset for a Real Time Device, an Edge Device, or a System", model.TypeId)
	}

	if model.TypeId == structs.AssetModelTypeIdSystem {
		if hwSerial != "" {
			return fmt.Errorf("for asset type %q the serial will be generated automatically", model.TypeId)
		}
		hwSerial = uuid.NewString()
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
