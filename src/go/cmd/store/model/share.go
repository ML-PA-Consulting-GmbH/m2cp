package model

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share [modelId | modelType modelName modelRevision]",
	Short: "Make a model globally shared across all tenants (irreversible)",
	Args:  validateModelShareArgs,
	RunE:  runShareCmd,
}

func init() {
	ModelCmd.AddCommand(shareCmd)
}

func validateModelShareArgs(cmd *cobra.Command, args []string) error {
	return validateModelArgs(cmd, args)
}

func runShareCmd(cmd *cobra.Command, args []string) error {
	revisionId, err := modelIdFromArgs(cmd, args)
	if err != nil {
		return err
	}

	deviceModelId, err := backend.GetDeviceModelIdByRevisionId(cmd.Context(), string(revisionId))
	if err != nil {
		return err
	}

	result, err := backend.ShareDeviceModel(cmd.Context(), deviceModelId)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, result, customModelShareFormatter)
}

func customModelShareFormatter(result *backend.ShareDeviceModelResponse) (string, error) {
	if result == nil || len(result.ShareDeviceModels) == 0 {
		return "", fmt.Errorf("no model was shared")
	}

	return fmt.Sprintf("Model %s is now globally shared", result.ShareDeviceModels[0].Id), nil
}
