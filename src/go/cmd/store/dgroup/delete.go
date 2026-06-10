package dgroup

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [deploymentGroupName or Id]",
	Short: "Delete a Deployment Group",
	Long:  "Delete a Deployment Group. Must be empty - containing no Devices",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeleteCmd,
}

type StoreDGroupDeleteResult struct {
	Message                string      `json:"message"`
	DeletedDeploymentGroup interface{} `json:"deletedDeploymentGroup"`
}

func init() {
	deleteCmd.Flags().Bool("force", true, "enforce to unset the contained devices")
	FleetCmd.AddCommand(deleteCmd)
	DGroupCmd.AddCommand(deleteCmd)
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	var deploymentGroupId string

	// Check if the argument is a UUID or a name
	if tools.IsValidUuid(args[0]) {
		deploymentGroupId = args[0]
	} else {
		// Look up deployment group by name
		name := args[0]
		resp, err := backend.GetDeploymentGroupIdByName(cmd.Context(), name)
		if err != nil {
			return err
		}

		if resp.DeploymentGroups == nil || resp.DeploymentGroups.Items == nil || len(resp.DeploymentGroups.Items) != 1 {
			return fmt.Errorf("deployment Group with name \"%s\" not found", name)
		}

		deploymentGroupId = resp.DeploymentGroups.Items[0].Id
	}

	if cmd.Flags().Changed("force") {
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}
		if force {
			dgroup, err := backend.GetDeploymentGroupByNameOrId(cmd.Context(), deploymentGroupId)
			if err != nil {
				return err
			}
			devices := dgroup.Devices
			for _, device := range devices {
				_, unsetErr := backend.UnsetDeviceDeploymentGroup(cmd.Context(), device.DeviceId)
				if unsetErr != nil {
					return unsetErr
				}
			}
		}
	}

	// Delete the deployment group using the new mutation
	deleteResult, err := backend.DeleteDeploymentGroups(cmd.Context(), []string{deploymentGroupId})
	if err != nil {
		return err
	}

	if len(deleteResult.DeleteDeploymentGroups) != 1 {
		return fmt.Errorf("failed to delete deployment group")
	}

	msg := StoreDGroupDeleteResult{
		Message:                "deployment group deleted",
		DeletedDeploymentGroup: deleteResult.DeleteDeploymentGroups[0],
	}

	return format.PrintFormattedOutput(cmd, msg, customStoreDeploymentGroupDeleteFormatter)
}

func customStoreDeploymentGroupDeleteFormatter(res StoreDGroupDeleteResult) (string, error) {
	return res.Message, nil
}
