package asset

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink <system-asset-id> <device-asset-id>",
	Short: "Remove a device asset from a system asset",
	Args:  cobra.ExactArgs(2),
	RunE:  runUnlinkCmd,
}

func init() {
	AssetCmd.AddCommand(unlinkCmd)
}

func runUnlinkCmd(cmd *cobra.Command, args []string) error {
	systemAssetId := args[0]
	deviceAssetId := args[1]

	// validate system asset
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

	// validate device asset
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

	// validate device asset is part of system
	systemBefore, err := backend.GetSystemById(cmd.Context(), systemAssetId)
	if err != nil {
		return err
	}
	if systemBefore == nil {
		return fmt.Errorf("asset not found")
	}
	found := false
	for _, childAsset := range systemBefore.Components {
		if childAsset.Id == deviceAssetId {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("device asset %s is not part of system %s", deviceAssetId, systemAssetId)
	}

	// update system
	if err = backend.SetAssetSystem(cmd.Context(), deviceAssetId, nil); err != nil {
		return fmt.Errorf("unable to remove asset '%s' from system: %s", deviceAssetId, err)
	}

	// print final system
	systemFinal, err := backend.GetSystemById(cmd.Context(), systemAssetId)
	if err != nil {
		return err
	}
	if systemFinal == nil {
		return fmt.Errorf("asset not found")
	}

	return format.PrintFormattedOutput(cmd, systemFinal, systemInfoFormatter)
}
