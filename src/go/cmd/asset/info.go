package asset

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"m2cpcli/tools/console"

	"github.com/spf13/cobra"
)

const systemFlagName = "system"

var infoCmd = &cobra.Command{
	Use:   "info <id>",
	Short: "info on a asset model",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfoCmd,
}

func init() {
	AssetCmd.AddCommand(infoCmd)
	infoCmd.Flags().Bool(systemFlagName, false, "show the full system of which the given asset is a part of")

}

func runInfoCmd(cmd *cobra.Command, args []string) error {
	id := args[0]

	// we first fetch the asset to find out about its system
	asset, err := backend.GetAssetById(cmd.Context(), id)
	if err != nil {
		return err
	}
	if asset == nil {
		return fmt.Errorf("asset not found")
	}

	if cmd.Flags().Changed(systemFlagName) {
		if asset.SystemAssetId == nil {
			return fmt.Errorf("asset is not part of a system")
		}
		systemAsset, err := backend.GetSystemById(cmd.Context(), *asset.SystemAssetId)
		if err != nil {
			return err
		}
		return format.PrintFormattedOutput(cmd, systemAsset, systemInfoFormatter)
	}
	return format.PrintFormattedOutput(cmd, asset, infoFormatter)

}

func systemInfoFormatter(systemAsset *structs.Asset) (string, error) {
	out := FormatSystem(systemAsset)
	return out, nil
}

func infoFormatter(asset *structs.Asset) (string, error) {
	out := FormatAsset(asset, nil)
	return out, nil
}

func FormatAsset(asset *structs.Asset, title *string) string {
	if title == nil {
		title = tools.Ptr("Asset")
	}
	out := ""
	out += fmt.Sprintf("%s:\n", console.Colorize(console.Green, *title))
	out += fmt.Sprintf("  ├─ Asset Id:                 %s\n", asset.Id)
	out += fmt.Sprintf("  ├─ Parent Asset Id:          %s\n", tools.MaybeStringToString(asset.ParentAssetId, "n/a"))
	out += fmt.Sprintf("  ├─ System Asset Id:          %s\n", tools.MaybeStringToString(asset.SystemAssetId, "n/a"))
	out += fmt.Sprintf("  ├─ Asset HW Serial:          %s\n", asset.Serial)
	if tools.MaybeStringToString(asset.AssetType, "") == structs.AssetTypeRtd {
		out += fmt.Sprintf("  ├─ Asset HW MCU Id:          %s\n", tools.MaybeStringToString(asset.McuId, "n/a"))
	}
	out += fmt.Sprintf("  ├─ Asset Type                %s\n", tools.MaybeStringToString(asset.AssetType, "n/a"))
	// we don't need these three, they are condensed into AssetType
	//out += fmt.Sprintf("  ├─ Asset Model Type Id       %s\n", asset.AssetModelTypeId)
	//out += fmt.Sprintf("  ├─ Asset Model Type Name     %s\n", asset.AssetModelTypeName)
	//out += fmt.Sprintf("  ├─ Asset Model Type Category %s\n", asset.AssetModelTypeCategory)
	out += fmt.Sprintf("  ├─ Asset Model Id            %s\n", asset.AssetModelId)
	out += fmt.Sprintf("  ├─ Asset Model Name          %s\n", asset.AssetModelName)
	out += fmt.Sprintf("  ├─ Asset Name:               %s\n", asset.Name)
	out += fmt.Sprintf("  ├─ Asset Description:        %s\n", tools.MaybeStringToString(asset.Description, "n/a"))
	out += fmt.Sprintf("  ├─ Device Id                 %s\n", tools.MaybeStringToString(asset.DeviceId, "n/a"))
	if asset.Device != nil {
		out += fmt.Sprintf("  ├─ Device OS Serial          %s\n", asset.Device.DeviceSerial)
		out += fmt.Sprintf("  ├─ Attestation Key           %s\n", tools.MaybeStringToString(asset.Device.DeviceAttestationKey, "n/a"))
	}
	out += fmt.Sprintf("  └─ Tenant                    %s (%s)\n", asset.TenantName, asset.TenantId)
	return out
}

func FormatSystem(asset *structs.Asset) string {
	out := FormatAsset(asset, tools.Ptr("System Asset"))

	if asset.Components == nil || len(asset.Components) == 0 {
		out += fmt.Sprintf("\n%s", console.Colorize(console.Yellow, "* no Components registered for this System"))
	} else {
		for _, childAsset := range asset.Components {
			out += "\n" + FormatAsset(&childAsset, tools.Ptr("Child Asset"))
		}
	}
	out += "\n"

	return out
}
