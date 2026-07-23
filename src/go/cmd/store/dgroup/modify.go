package dgroup

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:   "modify [name or Id]",
	Short: "Modify a Deployment Group",
	Args:  cobra.ExactArgs(1), // TODO: maybe validate with RegExp?
	RunE:  runModifyCmd,
}

type DeploymentGroupModifyResult struct {
	Message                 string      `json:"message"`
	ModifiedDeploymentGroup interface{} `json:"modifiedDeploymentGroup"`
}

func init() {
	modifyCmd.Flags().StringP("name", "n", "", "new name of the Deployment Group")
	modifyCmd.Flags().StringP("description", "d", "", "new description of the Deployment Group")
	modifyCmd.Flags().StringP("auto-update-mode", "m", "", "auto update mode (off, stable, edge)")
	modifyCmd.Flags().String("delta-updates-only", "", "restrict the Deployment Group to delta updates only (true or false)")
	FleetCmd.AddCommand(modifyCmd)
	DGroupCmd.AddCommand(modifyCmd)
}

func runModifyCmd(cmd *cobra.Command, args []string) error {

	var dgroupId string

	if tools.IsValidUuid(args[0]) {
		dgroupId = args[0]
	} else {
		resp, err := backend.GetFleetIdByName(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		if resp.Fleets == nil || resp.Fleets.Items == nil || len(resp.Fleets.Items) == 0 {
			return fmt.Errorf("Deployment Group with name %s not found", args[0])
		}

		dgroupId = resp.Fleets.Items[0].Id
	}

	// Only send the fields the user actually provided. The mutation input has
	// omitempty, so nil pointers are omitted rather than sent as empty values -
	// important for autoUpdateModeId, whose UUID type the backend cannot parse
	// from an empty string.
	var name *string
	if cmd.Flags().Changed("name") {
		v, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		name = &v
	}

	var description *string
	if cmd.Flags().Changed("description") {
		v, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		description = &v
	}

	var autoUpdateModeId *string
	if cmd.Flags().Changed("auto-update-mode") {
		autoUpdateModeName, err := cmd.Flags().GetString("auto-update-mode")
		if err != nil {
			return err
		}
		modeId, found := backend.GetAutoUpdateModeIdByName(autoUpdateModeName)
		if !found {
			return fmt.Errorf("unknown auto update mode '%s'", autoUpdateModeName)
		}
		autoUpdateModeId = &modeId
	}

	var isDeltaUpdateOnly *bool
	if cmd.Flags().Changed("delta-updates-only") {
		raw, err := cmd.Flags().GetString("delta-updates-only")
		if err != nil {
			return err
		}
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid value %q for --delta-updates-only: expected true or false", raw)
		}
		isDeltaUpdateOnly = &value
	}

	if name == nil && description == nil && autoUpdateModeId == nil && isDeltaUpdateOnly == nil {
		return fmt.Errorf("nothing to modify: provide at least one of --name, --description, --auto-update-mode, --delta-updates-only")
	}

	resp, err := backend.UpdateDeploymentGroup(cmd.Context(), dgroupId, name, description, autoUpdateModeId, isDeltaUpdateOnly)
	if err != nil {
		return err
	}

	msg := DeploymentGroupModifyResult{
		Message:                 "modified",
		ModifiedDeploymentGroup: resp.UpdateDeploymentGroups,
	}

	return format.PrintFormattedOutput(cmd, msg, customDeploymentGroupModifyFormatter)
}

func customDeploymentGroupModifyFormatter(res DeploymentGroupModifyResult) (string, error) {
	return res.Message, nil
}
