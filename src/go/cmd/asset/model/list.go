package model

import (
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List asset models of Edge Devices and Real Time Devices",
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

func init() {
	listCmd.Flags().Bool("ed", true, "show models for Edge Devices")
	listCmd.Flags().Bool("rtd", true, "show models for Real Time Devices")
	listCmd.Flags().Bool("sys", true, "show models for Systems")
	AssetModelCmd.AddCommand(listCmd)

}

func runListCmd(cmd *cobra.Command, args []string) error {
	types := []string{}
	if cmd.Flags().Changed("ed") {
		types = append(types, structs.AssetModelTypeIdEd)
	}
	if cmd.Flags().Changed("rtd") {
		types = append(types, structs.AssetModelTypeIdRtd)
	}
	if cmd.Flags().Changed("sys") {
		types = append(types, structs.AssetModelTypeIdSystem)
	}
	if len(types) == 0 {
		types = []string{structs.AssetModelTypeIdEd, structs.AssetModelTypeIdRtd, structs.AssetModelTypeIdSystem}
	}

	models, err := backend.GetAssetModels(cmd.Context(), types)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, models, modelsFormatter)
}

func modelsFormatter(models []structs.AssetModel) (string, error) {
	out := ""

	table := format.NewTable(map[string]string{
		"id":          "Id",
		"type":        "Type",
		"name":        "Name",
		"revision":    "Revision",
		"description": "Description",
	})

	for _, model := range models {
		table.AddRow(map[string]string{
			"id":          model.Id,
			"type id":     model.TypeId,
			"name":        model.Name,
			"revision":    tools.MaybeIntToString(model.Revision, "n/a"),
			"description": tools.MaybeStringToString(model.Description, "n/a"),
		})
	}
	out += table.StringSelect([]string{"id", "type", "name", "revision", "description"})

	return out, nil
}
