package app

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"
	"strings"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share <appId | appName> <arch>",
	Short: "Make an app globally shared across all tenants (irreversible)",
	Args:  validateAppShareArgs,
	RunE:  runShareCmd,
}

func init() {
	AppCmd.AddCommand(shareCmd)
}

func validateAppShareArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
		return err
	}

	switch len(args) {
	case 1:
		appId := args[0]
		if !tools.IsValidUuid(appId) {
			return fmt.Errorf("invalid appId '%s'", appId)
		}
	case 2:
		appName := args[0]
		arch := args[1]
		if !tools.IsValidAppName(appName) {
			return fmt.Errorf("invalid appName '%s'", appName)
		}
		if !tools.IsValidDeviceArchitecture(arch) {
			return fmt.Errorf("invalid Architecture '%s'", arch)
		}
	default:
		return fmt.Errorf("too many arguments")
	}

	return nil
}

func runShareCmd(cmd *cobra.Command, args []string) error {
	appId := args[0]

	if len(args) == 2 {
		appName := args[0]
		arch := strings.ToUpper(args[1])
		appIds, err := backend.GetAppIdsByAppNameAndArchitecture(cmd.Context(), appName, backend.Architecture(arch))
		if err != nil {
			return err
		}

		if len(appIds.Apps.Items) == 0 {
			return fmt.Errorf("no app found with name '%s' and architecture '%s'", appName, arch)
		}

		if len(appIds.Apps.Items) > 1 {
			return fmt.Errorf("multiple apps found with name '%s' and architecture '%s', please specify appId", appName, arch)
		}

		appId = appIds.Apps.Items[0].Id
	}

	result, err := backend.ShareApp(cmd.Context(), appId)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, result, customAppShareFormatter)
}

func customAppShareFormatter(result *backend.ShareAppResponse) (string, error) {
	if result == nil || len(result.ShareApps) == 0 {
		return "", fmt.Errorf("no app was shared")
	}

	return fmt.Sprintf("App %s is now globally shared", result.ShareApps[0].Id), nil
}
