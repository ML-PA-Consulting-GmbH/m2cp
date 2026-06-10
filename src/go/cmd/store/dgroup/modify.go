package dgroup

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:   "modify [name or Id]",
	Short: "Modify a Deployment Group",
	Args:  cobra.ExactArgs(1), // TODO: maybe validate with RegExp?
	RunE:  runModifyCmd,
}

type FleetModifyResult struct {
	Message       string      `json:"message"`
	ModifiedFleet interface{} `json:"modifiedFleet"`
}

func init() {
	modifyCmd.Flags().StringP("name", "n", "", "new name of the Deployment Group")
	modifyCmd.Flags().StringP("description", "d", "", "new description of the Deployment Group")
	modifyCmd.Flags().StringP("auto-update-mode", "m", "", "auto update mode (off, stable, edge)")
	FleetCmd.AddCommand(modifyCmd)
	DGroupCmd.AddCommand(modifyCmd)
}

func runModifyCmd(cmd *cobra.Command, args []string) error {

	var fleetId string

	if tools.IsValidUuid(args[0]) {
		fleetId = args[0]
	} else {
		resp, err := backend.GetFleetIdByName(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		if resp.Fleets == nil || resp.Fleets.Items == nil || len(resp.Fleets.Items) == 0 {
			return fmt.Errorf("Deployment Group with name %s not found", args[0])
		}

		fleetId = resp.Fleets.Items[0].Id
	}

	fleetName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	autoUpdateModeName, err := cmd.Flags().GetString("auto-update-mode")
	if err != nil {
		return err
	}

	var fleetAutoUpdateModeId string
	if autoUpdateModeName != "" {
		modeId, found := backend.GetAutoUpdateModeIdByName(autoUpdateModeName)
		if !found {
			return fmt.Errorf("unknown auto update mode '%s'", autoUpdateModeName)
		}
		fleetAutoUpdateModeId = modeId
	}

	resp, err := backend.UpdateFleet(cmd.Context(), fleetId, &fleetName, &description, &fleetAutoUpdateModeId)
	if err != nil {
		return err
	}

	msg := FleetModifyResult{
		Message:       "modified",
		ModifiedFleet: resp.UpdateFleets,
	}

	return format.PrintFormattedOutput(cmd, msg, customFleetModifyFormatter)
}

func customFleetModifyFormatter(res FleetModifyResult) (string, error) {
	return res.Message, nil
}
