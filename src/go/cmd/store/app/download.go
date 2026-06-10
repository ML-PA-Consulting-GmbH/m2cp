package app

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/helper"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <appId | appName> <arch> <revision | version>",
	Short: "Get information on an app in your store",
	Args:  cobra.ExactArgs(3),
	RunE:  runDownloadCmd,
}

func init() {
	AppCmd.AddCommand(downloadCmd)
	// Optional output flag
	downloadCmd.Flags().StringP("output", "o", "", "filepath to the output file (optionally)")
	// Optional boolean flag to disable progress bar
	downloadCmd.Flags().BoolP("no-progress-bar", "", false, "disable progress bar")
	// Optional flag to print the download URL instead of downloading the file
	downloadCmd.Flags().Bool("url", false, "print the download URL instead of downloading the file")
}

func runDownloadCmd(cmd *cobra.Command, args []string) error {
	appName := args[0]
	arch := strings.ToUpper(args[1])
	revisionOrVersion := args[2]

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

	appId := appIds.Apps.Items[0].Id

	// Auto-detect: if parsable as int, treat as revision number; otherwise as version string
	revisionInt, err := strconv.Atoi(revisionOrVersion)
	if err != nil {
		// Not a number — resolve version string to revision number
		appRevisions, err := backend.GetAppRevisionsByAppId(cmd.Context(), appId, 1000, 0)
		if err != nil {
			return err
		}
		found := false
		for _, item := range appRevisions.AppRevisions.Items {
			if item.Version == revisionOrVersion {
				revisionInt = item.Revision
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("version '%s' not found for app '%s' (%s)", revisionOrVersion, appName, arch)
		}
	}

	appInfo, err := backend.GetAppInfo(cmd.Context(), appId)

	appRevision, err := backend.GetAppRevision(cmd.Context(), appId, revisionInt)
	if err != nil {
		return err
	}

	data, err := backend.GetAppDownloadUrl(cmd.Context(), appRevision.Id)
	if err != nil {
		return err
	}
	downloadUrl := data.DownloadAppRevision.DownloadUrl

	urlOnly, _ := cmd.Flags().GetBool("url")

	result := AppDownloadResult{
		Revision:        appRevision.Revision,
		AppRevisionId:   appRevision.Id,
		AppName:         appInfo.App.AppName,
		AppArchitecture: string(appInfo.App.Architecture),
		AppId:           appInfo.App.Id,
	}

	if urlOnly {
		result.DownloadUrl = downloadUrl
	} else {
		output, _ := cmd.Flags().GetString("output")
		noProgressBar, _ := cmd.Flags().GetBool("no-progress-bar")
		fallbackOutputPath := helper.CreateFallbackOutputPath(cmd, args)
		outputPath, err := helper.DownloadSnap(downloadUrl, output, fallbackOutputPath, !noProgressBar && !format.JsonOutputMode)
		if err != nil {
			return err
		}
		result.OutputPath = outputPath
	}

	return format.PrintFormattedOutput(cmd, result, customAppDownloadFormatter)
}

type AppDownloadResult struct {
	OutputPath      string `json:"outputPath,omitempty"`
	DownloadUrl     string `json:"downloadUrl,omitempty"`
	AppId           string `json:"appId"`
	AppName         string `json:"appName"`
	AppArchitecture string `json:"appArchitecture"`
	AppRevisionId   string `json:"appRevisionId"`
	Revision        int    `json:"revision"`
}

func customAppDownloadFormatter(res AppDownloadResult) (string, error) {
	if res.DownloadUrl != "" {
		return res.DownloadUrl, nil
	}
	output := "Downloaded revision " + strconv.Itoa(res.Revision) +
		" of " + res.AppName + " for " + res.AppArchitecture + " to " + res.OutputPath
	return output, nil
}
