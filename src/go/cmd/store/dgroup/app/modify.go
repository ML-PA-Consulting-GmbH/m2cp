package app

import (
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var appModifyCmd = &cobra.Command{
	Use:   "modify [dgroupName|dgroupId appName|appId newVersion|newRevision]",
	Short: "Modify the revision of an App in a Deployment Group",
	Args:  validateAppAddOrModifyCmdArgs,
	RunE:  runAppModifyCmd,
}

type StoreDGroupAddModifyResult struct {
	Message        string `json:"message"`
	ModifiedBridge string `json:"modifiedFleetBridgeSnapRevision"`
}

func init() {
	fleetSnapCmd.AddCommand(appModifyCmd)
	dgroupAppCmd.AddCommand(appModifyCmd)
}

func runAppModifyCmd(cmd *cobra.Command, args []string) error {
	dgroupNameOrId := args[0]
	appNameOrId := args[1]
	revisionOrVersion := args[2]

	dgroupId, appId, appRevisionId, err := resolve(cmd.Context(), dgroupNameOrId, appNameOrId, revisionOrVersion)
	if err != nil {
		return err
	}

	bridgeId, err := backend.AddAppRevisionToDeploymentGroup(cmd.Context(), dgroupId, appId, appRevisionId, true)
	if err != nil {
		return err
	}

	msg := StoreDGroupAddModifyResult{
		Message:        "App revision modified in Deployment Group",
		ModifiedBridge: bridgeId,
	}

	return format.PrintFormattedOutput(cmd, msg, customStoreFleetSnapModifyFormatter)
}

func customStoreFleetSnapModifyFormatter(res StoreDGroupAddModifyResult) (string, error) {
	return res.Message, nil
}
