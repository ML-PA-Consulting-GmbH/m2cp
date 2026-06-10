package model

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <type> <name> [description]",
	Short: "Create a new Asset Model of type 'ed' (Edge Device), 'rtd' (Real Time Device) or 'sys' (System)",
	Args:  cobra.RangeArgs(2, 3),
	RunE:  runCreateCmd,
}

func init() {
	AssetModelCmd.AddCommand(createCmd)
}

func runCreateCmd(cmd *cobra.Command, args []string) error {
	modelType := args[0]
	name := args[1]
	var description *string
	if len(args) > 2 {
		description = &args[2]
	}

	modelTypeId, err := structs.AssetModelTypeToId(modelType)
	if err != nil {
		return fmt.Errorf("bad value for model type, use 'ed', 'rtd' or 'sys': %s", err)
	}

	modelId, err := backend.CreateAssetModel(cmd.Context(), modelTypeId, name, description)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, modelId, createdFormatter)

}

func createdFormatter(result string) (string, error) {
	return fmt.Sprintf("Created Asset Model with id %s", result), nil
}
