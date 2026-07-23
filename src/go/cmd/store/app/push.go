package app

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/helper"
	"m2cpcli/tools"
	"path/filepath"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [appFilePath uploadMessage]",
	Short: "Push an app to the store",
	Args:  cobra.ExactArgs(2),
	RunE:  runPushCmd,
}

type pushOutput struct {
	Version      string `json:"version"`
	Revision     int    `json:"revision"`
	RatingResult string `json:"ratingResult,omitempty"`
}

func init() {
	AppCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringP("rating", "r", "", "set rating of the app after pushing. Requires that description is set")
	pushCmd.Flags().StringP("description", "d", "", "set description of the app rating. Requires that rating is set")
}

func runPushCmd(cmd *cobra.Command, args []string) error {

	appFilePath := args[0]
	appFileName := filepath.Base(appFilePath)
	binaryBytes, err := tools.ReadLocalFile(appFilePath)
	if err != nil {
		return err
	}

	rating, ratingErr := cmd.Flags().GetString("rating")
	description, descriptionErr := cmd.Flags().GetString("description")
	if (ratingErr == nil && descriptionErr != nil) || (ratingErr != nil && descriptionErr == nil) {
		return fmt.Errorf("rating and description flags must be used together")
	}

	fmt.Printf("Initiating upload for '%s'\n", appFilePath)

	initiateResponse, err := backend.InitiateAppRevisionUpload(cmd.Context(), &backend.AppRevisionInitiateUploadInput{
		FileName: appFileName,
		FileSize: int64(len(binaryBytes)),
	})

	if err != nil {
		return err
	}

	fmt.Printf("Uploading %d bytes to %s\n", len(binaryBytes), initiateResponse.InitiateAppRevisionUpload.ContinuesToken)

	err = helper.UploadToAzureBlobStorage(cmd.Context(), binaryBytes, initiateResponse.InitiateAppRevisionUpload.UploadUrl)

	if err != nil {
		return err
	}

	fmt.Printf("Upload complete, finalizing...\n")

	completeResponse, err := backend.CompleteAppRevisionUpload(cmd.Context(), &backend.AppRevisionCompleteUploadInput{
		UploadMessage:  args[1],
		ContinuesToken: initiateResponse.InitiateAppRevisionUpload.ContinuesToken,
	})

	if err != nil {
		return err
	}

	output := pushOutput{
		Version:  completeResponse.CompleteAppRevisionUpload.Version,
		Revision: completeResponse.CompleteAppRevisionUpload.Revision,
	}

	if ratingErr == nil && descriptionErr == nil {
		err := backend.SetAppRevisionRating(cmd.Context(), completeResponse.CompleteAppRevisionUpload.Id, rating, description)
		if err != nil {
			output.RatingResult = fmt.Sprintf("rating to '%s' failed: %s", rating, err.Error())
		} else {
			output.RatingResult = fmt.Sprintf("rating to '%s' succeeded", rating)
		}
	}

	return format.PrintFormattedOutput(cmd, output, customAppPushFormatter)
}

func customAppPushFormatter(output pushOutput) (string, error) {
	return fmt.Sprintf("App pushed successfully. Version: %s, Revision: %d", output.Version, output.Revision), nil
}
