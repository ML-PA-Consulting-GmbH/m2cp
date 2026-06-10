package app

import (
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var appRemoveCmd = &cobra.Command{
	Use:   "remove [dgroupName|dgroupId appName|appId]",
	Short: "Remove an App from a Deployment Group",
	Args:  cobra.ExactArgs(2),
	RunE:  runAppRemoveCmd,
}

type StoreDgroupAppRemoveResult struct {
	Message       string `json:"message"`
	DeletedBridge string `json:"removedFleetBridgeSnapRevision"`
}

func init() {
	fleetSnapCmd.AddCommand(appRemoveCmd)
	dgroupAppCmd.AddCommand(appRemoveCmd)
}

func runAppRemoveCmd(cmd *cobra.Command, args []string) error {
	dgroupId, appId, err := resolveDGroupAndApp(cmd.Context(), args[0], args[1])
	if err != nil {
		return err
	}

	bridgeId, err := backend.RemoveAppFromDeploymentGroup(cmd.Context(), dgroupId, appId)
	if err != nil {
		return err
	}

	msg := StoreDgroupAppRemoveResult{
		Message:       "App removed to Deployment Group",
		DeletedBridge: bridgeId,
	}

	return format.PrintFormattedOutput(cmd, msg, customStoreDGroupAppRemoveFormatter)
}

func customStoreDGroupAppRemoveFormatter(res StoreDgroupAppRemoveResult) (string, error) {
	return res.Message, nil
}
