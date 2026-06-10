package snap

import (
	"fmt"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"m2cpcli/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [snapId | snapName snapDeviceArchitecture]",
	Short: "Download a snap from the store",
	Args:  validateDownloadArgs,
	RunE:  runDownloadCmd,
}

func init() {
	SnapCmd.AddCommand(downloadCmd)

	// Optional int flag revision
	downloadCmd.Flags().StringP("revision", "r", "latest", "snap revision")
	// Optional output flag
	downloadCmd.Flags().StringP("output", "o", "", "filepath to the output file (optionally)")
	// Optional boolean flag to disable progress bar
	downloadCmd.Flags().BoolP("no-progress-bar", "", false, "disable progress bar")
}

func validateDownloadArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
		return err
	}
	switch len(args) {
	case 1:
		// only appId is expected
		snapId := args[0]
		if !tools.IsValidAppId(snapId) {
			return fmt.Errorf("invalid snapId '%s'", snapId)
		}
	case 2:
		// snapName and snapDeviceArchitecture are expected
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

func runDownloadCmd(cmd *cobra.Command, args []string) error {
	declarationId, err := retrieveSnapDatabaseIdByArgs(cmd, args)
	if err != nil {
		return err
	}

	snapInfo := SnapInfoResult{}
	snapInfo.SnapDeclaration, err = gql.SnapDeclarationByDatabaseId(cmd.Context(), declarationId, true, true)
	if err != nil {
		return err
	}

	// Find requested revision
	revisionArgument, _ := cmd.Flags().GetString("revision")
	revision, err := helper.FindRequestedRevision(revisionArgument, snapInfo.SnapDeclaration.SnapRevisions)
	if err != nil {
		return err
	}

	// Send request to get download url
	downloadUrl, err := gql.SnapDownloadUrlBySnapRevisionId(cmd.Context(), revision.Id)
	if err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	noProgressBar, _ := cmd.Flags().GetBool("no-progress-bar")
	fallbackOutputPath := helper.CreateFallbackOutputPath(cmd, args)
	outputPath, err := helper.DownloadSnap(downloadUrl, output, fallbackOutputPath, !noProgressBar && !format.JsonOutputMode)
	if err != nil {
		return err
	}

	result := SnapDownloadResult{
		OutputPath:       outputPath,
		Revision:         int(revision.Revision),
		SnapRevisionId:   string(revision.Id),
		SnapName:         snapInfo.SnapDeclaration.SnapName,
		SnapArchitecture: snapInfo.SnapDeclaration.SnapDeviceArchitecture,
		SnapId:           snapInfo.SnapDeclaration.SnapId,
	}
	return format.PrintFormattedOutput(cmd, result, customSnapDownloadFormatter)
}

type SnapDownloadResult struct {
	OutputPath       string `json:"outputPath"`
	SnapId           string `json:"snapId"`
	SnapName         string `json:"snapName"`
	SnapArchitecture string `json:"snapArchitecture"`
	SnapRevisionId   string `json:"snapRevisionId"`
	Revision         int    `json:"revision"`
}

func customSnapDownloadFormatter(res SnapDownloadResult) (string, error) {
	output := "Downloaded revision " + strconv.Itoa(res.Revision) +
		" of " + res.SnapName + " for " + res.SnapArchitecture + " to " + res.OutputPath
	return output, nil
}
