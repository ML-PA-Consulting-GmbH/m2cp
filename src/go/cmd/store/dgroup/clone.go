package dgroup

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone <sourceName> <targetName>",
	Short: "Clone a Deployment Group",
	Long:  `Clone a Deployment Group. The source must exist, the target must not exist.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runCloneCmd,
}

type cloneOutput struct {
	Message string `json:"message"`
	Id      string `json:"id"`
}

func init() {
	cloneCmd.Flags().StringP("description", "d", "", "description of the new Deployment Group")
	FleetCmd.AddCommand(cloneCmd)
	DGroupCmd.AddCommand(cloneCmd)
}

func runCloneCmd(cmd *cobra.Command, args []string) error {
	sourceName := args[0]
	targetName := args[1]
	description, _ := cmd.Flags().GetString("description")

	if sourceName == "" {
		return fmt.Errorf("source Deployment Group name is required")
	}

	if targetName == "" {
		return fmt.Errorf("target Deployment Group name is required")
	}

	if sourceName == targetName {
		return fmt.Errorf("source and target Deployment Group names must be different")
	}

	sourceDeploymentGroup, err := backend.GetDeploymentGroupByNameOrId(cmd.Context(), sourceName)
	if err != nil {
		return fmt.Errorf("failed to find source deployment group '%s': %w", sourceName, err)
	}

	// Prepare the clone input
	cloneInput := &backend.DeploymentGroupCloneInput{
		Id:   sourceDeploymentGroup.Id,
		Name: targetName,
	}

	if description != "" {
		cloneInput.Description = &description
	}

	// Clone the deployment group using the new mutation
	cloneResponse, err := backend.DeploymentGroupClone(cmd.Context(), cloneInput)
	if err != nil {
		return fmt.Errorf("failed to clone deployment group: %w", err)
	}

	output := cloneOutput{
		Message: "Deployment Group cloned successfully",
		Id:      cloneResponse.CloneDeploymentGroup.Id,
	}

	return format.PrintFormattedOutput(cmd, output, cloneOutputFormatter)
}

func cloneOutputFormatter(output cloneOutput) (string, error) {
	return output.Message, nil
}
