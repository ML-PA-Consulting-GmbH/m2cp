package rate

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [snapName snapArchitecture revision rating description]",
	Short: "Set the rating of a snap in the store (unrated, experimental, stable, edge, deprecated, broken, denied)",
	Args:  validateSetSnapRatingArgs,
	RunE:  runSetCmd,
}

func init() {
	rateCmd.AddCommand(setCmd)
}

func validateSetSnapRatingArgs(cmd *cobra.Command, args []string) error {
	err := cobra.ExactArgs(5)(cmd, args)
	if err != nil {
		return err
	}

	if !tools.IsValidAppName(args[0]) {
		return fmt.Errorf("invalid snapName '%s'", args[0])
	}

	if !tools.IsValidDeviceArchitecture(args[1]) {
		return fmt.Errorf("invalid snapDeviceArchitecture '%s'", args[1])
	}

	if args[2] == "latest" {
		return fmt.Errorf("revision latest is not allowed, please use a specific revision")
	}
	if !tools.IsValidRevision(args[2]) {
		return fmt.Errorf("invalid snapRevision '%s'", args[2])
	}

	return nil
}

func runSetCmd(cmd *cobra.Command, args []string) error {
	appName := args[0]
	appArch := args[1]
	revisionStr := args[2]
	description := args[4]

	snapStatusId, found := backend.GetAppStatusIdByName(args[3])
	if !found {
		return fmt.Errorf("unknown rating '%s'", args[3])
	}

	app, err := backend.GetAppByNameAndArch(cmd.Context(), appName, appArch)
	if err != nil {
		return err
	}

	revision, err := strconv.Atoi(revisionStr)
	if err != nil {
		return err
	}
	res, err := backend.GetAppRevision(cmd.Context(), app.Id, revision)
	if err != nil {
		return err
	}
	revisionId := res.Id

	_, err = backend.UpdateSnapRevisionStatus(cmd.Context(), &backend.SnapRevisionUpdateStatusInput{
		Id:           revisionId,
		SnapStatusId: snapStatusId,
		Description:  description,
	})

	if err != nil {
		return err
	}

	cmd.Println("Rating updated")
	return nil
}
