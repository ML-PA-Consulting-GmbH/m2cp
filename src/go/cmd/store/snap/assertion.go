package snap

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"m2cpcli/tools"
	"os"
)

var assertionCmd = &cobra.Command{
	Use:   "assertion [snapId | snapName snapDeviceArchitecture]",
	Short: "Get assertions of a snap in your store",
	Args:  validateSnapAssertionArgs,
	RunE:  runAssertionCmd,
}

func init() {
	SnapCmd.AddCommand(assertionCmd)
	assertionCmd.Flags().StringP("revision", "r", "latest", "request specific revision number")
	var err error
	err = viper.BindPFlag("revision", assertionCmd.Flags().Lookup("revision"))
	cobra.CheckErr(err)
}

func validateSnapAssertionArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
		return err
	}

	switch len(args) {
	case 1:
		snapId := args[0]
		if !tools.IsValidAppId(snapId) {
			return fmt.Errorf("invalid snapId '%s'", snapId)
		}
	case 2:
		// snapName and snapDeviceArchitecture are given
		snapName := args[0]
		snapDeviceArchitecture := args[1]
		if !tools.IsValidAppName(snapName) {
			return fmt.Errorf("invalid snapName '%s'", snapName)
		}
		if !tools.IsValidDeviceArchitecture(snapDeviceArchitecture) {
			return fmt.Errorf("invalid snapDeviceArchitecture '%s'", snapDeviceArchitecture)
		}
	default:
		return fmt.Errorf("too many arguments")
	}

	// Validate revision
	revision, _ := cmd.Flags().GetString("revision")
	if !tools.IsValidRevision(revision) {
		return fmt.Errorf("invalid revision '%s'", revision)
	}

	return nil
}

func retrieveSnapDatabaseIdByArgs(cmd *cobra.Command, args []string) (gql.UUID, error) {
	if len(args) == 1 {
		return gql.SnapDatabaseIdBySnapId(cmd.Context(), args[0])
	} else {
		return gql.SnapDatabaseIdByNameAndArchitecture(cmd.Context(), args[0], args[1])
	}
}

func runAssertionCmd(cmd *cobra.Command, args []string) error {
	var err error
	id, err := retrieveSnapDatabaseIdByArgs(cmd, args)
	if err != nil {
		return err
	}

	snapInfo := SnapInfoResult{}
	snapInfo.SnapDeclaration, err = gql.SnapDeclarationByDatabaseId(cmd.Context(), id, true, true)
	if err != nil {
		return err
	}

	// Find requested revision
	revisionArgument, _ := cmd.Flags().GetString("revision")
	revision, err := helper.FindRequestedRevision(revisionArgument, snapInfo.SnapDeclaration.SnapRevisions)
	if err != nil {
		return err
	}

	assertions, err := gql.SnapAssertionsBySnapDatabaseIdAndRevision(cmd.Context(), snapInfo.SnapDeclaration.Id, revision.Revision)
	if err != nil {
		return err
	}

	formattedAssertions := helper.FormatAssertions(assertions)
	cmd.SetOut(os.Stdout)
	cmd.Print(formattedAssertions)
	return nil
}
