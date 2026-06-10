package model

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/cmd/store"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"strconv"
)

var ModelCmd = &cobra.Command{
	Use:     "model",
	Aliases: []string{"m"},
	Short:   "Manage the model",
}

func init() {
	store.StoreCmd.AddCommand(ModelCmd)
}

func validateModelArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.RangeArgs(1, 3)(cmd, args); err != nil {
		return err
	}

	switch len(args) {
	case 1:
		modelId := args[0]
		if !tools.IsValidUuid(modelId) {
			return fmt.Errorf("invalid model UUID '%s'", modelId)
		}
	case 3:
		//modelType := args[0]
		//modelName := args[1]
		modelRevision := args[2]
		if !tools.IsValidRevision(modelRevision) {
			return fmt.Errorf("invalid model revision '%s'", modelRevision)
		}
	default:
		return fmt.Errorf("either model ID or model type, name, and revision are required")
	}
	return nil
}

func modelIdFromArgs(cmd *cobra.Command, args []string) (gql.UUID, error) {
	var modelId gql.UUID

	switch len(args) {
	case 1:
		modelId = gql.UUID(args[0])
	case 3:
		modelType := args[0]
		modelName := args[1]
		// TODO: handle "latest
		modelRevision, err := strconv.Atoi(args[2])
		if err != nil {
			return "", fmt.Errorf("invalid model revision number: \"%s\": %s", args[2], err)
		}
		modelId, err = gql.EdgeDeviceModelIdByTypeNameRevision(cmd.Context(), modelType, modelName, modelRevision)
		if err != nil {
			return "", fmt.Errorf("could not find model by Type, Name, and Revision (%s, %s, %d): %s",
				modelType, modelName, modelRevision, err)
		}
	default:
		return "", fmt.Errorf("model id or model Type, Name, and Revision is required")
	}
	return modelId, nil
}
