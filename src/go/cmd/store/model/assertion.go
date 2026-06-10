package model

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
)

var assertionCmd = &cobra.Command{
	Use:   "assertion [modelId | modelType modelName modelRevision]",
	Short: "Get the assertion of a model in your store",
	Args:  validateModelAssertionArgs,
	RunE:  runAssertionCmd,
}

func validateModelAssertionArgs(cmd *cobra.Command, args []string) error {
	return validateModelArgs(cmd, args)
}

func init() {
	ModelCmd.AddCommand(assertionCmd)
}

type ModelAssertionResult struct {
	Assertion string `json:"assertion"`
}

func runAssertionCmd(cmd *cobra.Command, args []string) error {
	modelId, err := modelIdFromArgs(cmd, args)
	if err != nil {
		return err
	}

	var model *gql.EdgeDeviceModel
	model, err = gql.EdgeDeviceModelById(cmd.Context(), modelId, gql.EdgeDeviceModelQueryOptions{WithAssertion: true})
	if err != nil {
		return fmt.Errorf("could not find model by ID \"%s\": %s", modelId, err)
	}

	modelAssertionResult := ModelAssertionResult{
		Assertion: model.EdgeDeviceModelAssertion.AssertionBody,
	}

	return format.PrintFormattedOutput(cmd, modelAssertionResult, customModelAssertionFormatter)
}

func customModelAssertionFormatter(res ModelAssertionResult) (string, error) {
	return res.Assertion, nil
}
