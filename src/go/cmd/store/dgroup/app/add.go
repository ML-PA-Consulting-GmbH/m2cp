package app

import (
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var appAddCmd = &cobra.Command{
	Use:   "add [dgroupName|dgroupId appName|appId newVersion|newRevision]",
	Short: "Add an App to a Deployment Group",
	Long:  `Add an App of a specific revision to a Deployment Group`,
	Args:  validateAppAddOrModifyCmdArgs,
	RunE:  runAppAddCmd,
}

type StoreDGroupAppAddResult struct {
	Message string `json:"message"`
	Bridge  string `json:"createdFleetBridgeSnapRevision"`
}

func init() {
	appAddCmd.Flags().StringP("revision", "r", "latest", "revision of the App (default: latest)")
	fleetSnapCmd.AddCommand(appAddCmd)
	dgroupAppCmd.AddCommand(appAddCmd)
}

func runAppAddCmd(cmd *cobra.Command, args []string) error {
	dgroupNameOrId := args[0]
	appNameOrId := args[1]
	revisionOrVersion := args[2]

	dgroupId, appId, appRevisionId, err := resolve(cmd.Context(), dgroupNameOrId, appNameOrId, revisionOrVersion)
	if err != nil {
		return err
	}

	bridgeId, err := backend.AddAppRevisionToDeploymentGroup(cmd.Context(), dgroupId, appId, appRevisionId, false)
	if err != nil {
		return err
	}

	msg := StoreDGroupAppAddResult{
		Message: "App added to Deployment Group",
		Bridge:  bridgeId,
	}

	return format.PrintFormattedOutput(cmd, msg, customStoreFleetSnapAddFormatter)
}

func customStoreFleetSnapAddFormatter(res StoreDGroupAppAddResult) (string, error) {
	return res.Message, nil
}
