package rate

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [appName appArchitecture revision rating description]",
	Short: "Set the rating of an App in the App Store (unrated, experimental, stable, edge, deprecated, broken, denied)",
	Args:  validateSetAppRatingArgs,
	RunE:  runSetCmd,
}

func init() {
	rateCmd.AddCommand(setCmd)
}

func validateSetAppRatingArgs(cmd *cobra.Command, args []string) error {
	err := cobra.ExactArgs(5)(cmd, args)
	if err != nil {
		return err
	}

	if !tools.IsValidAppName(args[0]) {
		return fmt.Errorf("invalid appName '%s'", args[0])
	}

	if !tools.IsValidDeviceArchitecture(args[1]) {
		return fmt.Errorf("invalid appDeviceArchitecture '%s'", args[1])
	}

	if args[2] == "latest" {
		return fmt.Errorf("revision latest is not allowed, please use a specific revision")
	}
	if !tools.IsValidRevision(args[2]) {
		return fmt.Errorf("invalid appRevision '%s'", args[2])
	}

	return nil
}

func runSetCmd(cmd *cobra.Command, args []string) error {
	appNameOrId := args[0]
	appArchitecture := args[1]
	revisionStr := args[2]
	rating := args[3]
	description := args[4]

	app, err := backend.GetAppByNameAndArch(cmd.Context(), appNameOrId, appArchitecture)
	if err != nil {
		return err
	}

	revision, err := strconv.Atoi(revisionStr)
	if err != nil {
		return err
	}
	appRevision, err := backend.GetAppRevision(cmd.Context(), app.Id, revision)
	if err != nil {
		return err
	}

	err = backend.SetAppRevisionRating(cmd.Context(), appRevision.Id, rating, description)
	if err != nil {
		return err
	}

	cmd.Printf("Rating of App %s (%s) Version %s (%d) updated to %s\n", app.AppName, app.Architecture,
		appRevision.Version, appRevision.Revision, rating)
	return nil
}
