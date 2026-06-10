package model

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/backend"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"strconv"
	"strings"
)

var infoCmd = &cobra.Command{
	Use:   "info [modelId | modelType modelName modelRevision]",
	Short: "Get information on a model in your store",
	Args:  validateModelInfoArgs,
	RunE:  runInfoCmd,
}

func validateModelInfoArgs(cmd *cobra.Command, args []string) error {
	return validateModelArgs(cmd, args)
}

func init() {
	ModelCmd.AddCommand(infoCmd)
}

type SnapRevisionRow struct {
	RevisionId  gql.UUID `json:"revision-id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Revision    int32    `json:"revision"`
}

type InfoResult struct {
	DeviceModelRevision *backend.GetDeviceModelRevisionInfoByIdResponse `json:"device_model_revision,omitempty"`
}

func extractSnapDeclarations(bridge []gql.EdgeDeviceModelBridgeSnapDeclaration) []gql.SnapDeclaration {
	result := make([]gql.SnapDeclaration, len(bridge))
	for idx, b := range bridge {
		result[idx] = b.SnapDeclaration
	}
	return result
}

func runInfoCmd(cmd *cobra.Command, args []string) error {
	modelId, err := modelIdFromArgs(cmd, args)
	if err != nil {
		return err
	}

	deviceModelRevision, err := backend.GetDeviceModelRevisionInfoById(cmd.Context(), string(modelId))
	if err != nil {
		return err
	}

	modelInfoResult := InfoResult{
		DeviceModelRevision: deviceModelRevision,
	}

	return format.PrintFormattedOutput(cmd, modelInfoResult, customModelInfoFormatter)
}

func customModelInfoFormatter(res InfoResult) (string, error) {
	info := format.NewList()
	info.Add("Id", fmt.Sprintf("%s", res.DeviceModelRevision.DeviceModelRevision.Id))
	info.Add("Type", res.DeviceModelRevision.DeviceModelRevision.DeviceModel.DeviceType.DeviceTypeName)
	info.Add("Name", res.DeviceModelRevision.DeviceModelRevision.DeviceModel.ModelName)
	info.Add("Revision", fmt.Sprintf("%d", *res.DeviceModelRevision.DeviceModelRevision.Revision))
	info.Add("Architecture", strings.ToLower(string(res.DeviceModelRevision.DeviceModelRevision.DeviceModel.Architecture)))
	info.Add("TPM required", strconv.FormatBool(res.DeviceModelRevision.DeviceModelRevision.IsTpmRequired))
	info.Add("Pre-Registration required", strconv.FormatBool(res.DeviceModelRevision.DeviceModelRevision.IsPreRegistrationRequired))
	info.Add("Upload Message", res.DeviceModelRevision.DeviceModelRevision.UploadMessage)

	modelSnapsTable := format.NewTable(map[string]string{
		"1id":      "Id",
		"2name":    "Name",
		"3base":    "Base",
		"4summary": "Summary",
	})

	for _, decl := range res.DeviceModelRevision.DeviceModelRevision.DeviceModelRevisionBridgeApps {
		app := decl.App
		snapBase := ""
		if app.AppSnap != nil {
			snapBase = app.AppSnap.SnapBase
		}

		modelSnapsTable.AddRow(map[string]string{
			"1id":      app.Id,
			"2name":    app.AppName,
			"3base":    snapBase,
			"4summary": strings.TrimSpace(tools.ShortenRight(app.Summary, 45)),
		})
	}
	modelSnapsTable = modelSnapsTable.Sort("2name")

	return fmt.Sprintf("Model\n%s\n"+"Core Apps\n%s", info.String(), modelSnapsTable.String()), nil
}
