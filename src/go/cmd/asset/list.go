package asset

import (
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list systems",
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

func init() {
	listCmd.Flags().Bool("ed", true, "show models for Edge Devices")
	listCmd.Flags().Bool("rtd", true, "show models for Real Time Devices")
	listCmd.Flags().Bool("sys", true, "show models for Systems")
	AssetCmd.AddCommand(listCmd)
}

func runListCmd(cmd *cobra.Command, args []string) error {
	filter, err := getSystemListFilter(cmd)
	if err != nil {
		return err
	}

	systems, err := backend.GetSystemsList(cmd.Context(), filter)

	return format.PrintFormattedOutput(cmd, systems, assetListFormatter)
}

func assetListFormatter(assets []structs.Asset) (string, error) {
	out := ""

	table := format.NewTable(map[string]string{
		"id":       "Id",
		"hwSerial": "HwSerial",
		"hwMcuId":  "HwMcuId",
		"name":     "Name",
		"type":     "Type",
		"model":    "Model",
		"deviceId": "DeviceId",
	})

	for _, asset := range assets {
		table.AddRow(map[string]string{
			"id":       asset.Id,
			"hwSerial": asset.Serial,
			"hwMcuId":  tools.MaybeStringToString(asset.McuId, "n/a"),
			"name":     asset.Name,
			"type":     formatAssetType(asset.AssetModelTypeName, asset.AssetModelTypeCategory),
			"model":    asset.AssetModelName,
			"deviceId": tools.MaybeStringToString(asset.DeviceId, "n/a"),
		})
	}
	out += table.StringSelect([]string{"id", "type", "model", "hwSerial", "name", "deviceId"})

	return out, nil
}

type QueryFilter struct {
	Field      string
	Comparison string
	Value      string
}

func getSystemListFilter(cmd *cobra.Command) ([]structs.BackendQueryFilter, error) {
	filters := []structs.BackendQueryFilter{}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return nil, err
		}
		filters = append(filters, structs.BackendQueryFilter{
			Field:   structs.BackendQueryFilterAssetName,
			Pattern: &name,
		})
	}

	types := []string{}
	if cmd.Flags().Changed("ed") {
		types = append(types, structs.AssetTypeIdEd)
	}
	if cmd.Flags().Changed("rtd") {
		types = append(types, structs.AssetTypeIdRtd)
	}
	if cmd.Flags().Changed("sys") {
		types = append(types, structs.AssetTypeIdSystem)
	}
	if len(types) > 0 {
		filters = append(filters, structs.BackendQueryFilter{
			Field:    structs.BackendQueryFilterAssetModelTypeId,
			ExactAny: types,
		})
	}

	if cmd.Flags().Changed("sort") {
		sort, err := cmd.Flags().GetString("sort")
		if err != nil {
			return nil, err
		}
		switch sort {
		case "id":
			filters = append(filters, structs.BackendQueryFilter{
				Field: structs.BackendQueryFilterAssetId,
				Sort:  tools.Ptr(structs.BackendQueryFilterSortAsc),
			})
		case "created":
			filters = append(filters, structs.BackendQueryFilter{
				Field: structs.BackendQueryFilterCreatedAt,
				Sort:  tools.Ptr(structs.BackendQueryFilterSortAsc),
			})
		case "modified":
			filters = append(filters, structs.BackendQueryFilter{
				Field: structs.BackendQueryFilterModifiedAt,
				Sort:  tools.Ptr(structs.BackendQueryFilterSortAsc),
			})
		case "name":
		default:
			filters = append(filters, structs.BackendQueryFilter{
				Field: structs.BackendQueryFilterAssetName,
				Sort:  tools.Ptr(structs.BackendQueryFilterSortAsc),
			})
		}
	} else {
		filters = append(filters, structs.BackendQueryFilter{
			Field: structs.BackendQueryFilterCreatedAt,
			Sort:  tools.Ptr(structs.BackendQueryFilterSortAsc),
		})
	}

	return filters, nil
}

func formatAssetType(t, c string) string {
	if t == "System" {
		return "System"
	} else if t == "M2CP" {
		return "Edge Device"
	} else if t == "RTOS" {
		return "Real Time Device"
	}
	return "unknown"
}
