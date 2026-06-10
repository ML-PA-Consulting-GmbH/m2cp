package asset

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <system-asset-id> <device-asset-id>",
	Short: "Make an existing device asset a child of an existing system",
	Args:  cobra.ExactArgs(2),
	RunE:  runLinkCmd,
}

func init() {
	AssetCmd.AddCommand(linkCmd)
}

func runLinkCmd(cmd *cobra.Command, args []string) error {
	systemAssetId := args[0]
	deviceAssetId := args[1]

	systemAsset, err := backend.GetAssetById(cmd.Context(), systemAssetId)
	if err != nil {
		return err
	}
	if systemAsset == nil {
		return fmt.Errorf("system asset not found")
	}
	if tools.MaybeStringToString(systemAsset.AssetType, "") != structs.AssetTypeSystem {
		return fmt.Errorf("asset %s is not a System", systemAssetId)
	}

	deviceAsset, err := backend.GetAssetById(cmd.Context(), deviceAssetId)
	if err != nil {
		return err
	}
	if deviceAsset == nil {
		return fmt.Errorf("device asset not found")
	}
	if tools.MaybeStringToString(deviceAsset.AssetType, "") != structs.AssetTypeEd && tools.MaybeStringToString(deviceAsset.AssetType, "") != structs.AssetTypeRtd {
		return fmt.Errorf("asset %s is neither an Edge Device nor a Real Time Device", deviceAssetId)
	}

	if err = backend.SetAssetSystem(cmd.Context(), deviceAssetId, &systemAssetId); err != nil {
		return fmt.Errorf("unable to set asset '%s' to system: %s", deviceAssetId, err)
	}

	systemFinal, err := backend.GetSystemById(cmd.Context(), systemAssetId)
	if err != nil {
		return err
	}
	if systemFinal == nil {
		return fmt.Errorf("asset not found")
	}

	return format.PrintFormattedOutput(cmd, systemFinal, systemInfoFormatter)
}
