package snap

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [snapFilePath developerNote]",
	Short: "Push a snap to the store",
	Args:  validatePushCmdArgs,
	RunE:  runPushCmd,
}

func init() {
	SnapCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringP("rating", "r", "", "set rating of the snap after pushing. Requires that description is set")
	pushCmd.Flags().StringP("description", "d", "", "set description of the snap rating. Requires that rating is set")
}

func validatePushCmdArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(2)(cmd, args); err != nil {
		return err
	}

	// Check, that when rating is set, description is also set
	ratingInput := cmd.Flag("rating").Value.String()
	descriptionInput := cmd.Flag("description").Value.String()
	if ratingInput != "" && descriptionInput == "" {
		return fmt.Errorf("description must be set when rating is set")
	}

	// Check, that when description is set, rating is also set
	if descriptionInput != "" && ratingInput == "" {
		return fmt.Errorf("rating must be set when description is set")
	}

	return nil
}

type SnapPushResult struct {
	UploadMessage string             `json:"upload-message"`
	FileSize      int                `json:"size"`
	Result        gql.PushSnapOutput `json:"result"`
	Rating        string
}

func runPushCmd(cmd *cobra.Command, args []string) error {
	//Check first, whether the rating exists if we have to set it after pushing
	var ratingId string = ""
	ratingInput := cmd.Flag("rating").Value.String()
	descriptionInput := cmd.Flag("description").Value.String()

	if ratingInput != "" {
		statusId, found := backend.GetAppStatusIdByName(ratingInput)
		if !found {
			return fmt.Errorf("unknown snap rating '%s'", ratingInput)
		}
		ratingId = statusId
	}

	snapFilePath := args[0]
	binaryBytes, err := tools.ReadLocalFile(snapFilePath)
	if err != nil {
		return err
	}
	snapFileSize := len(binaryBytes)

	// Step 1: Generate upload url at backend
	uploadResponse, err := gql.GenerateSnapUploadUrl(cmd.Context(), snapFileSize)
	if err != nil {
		return err
	}

	// Step 2: Upload snap to Azure Blob Storage
	err = helper.UploadToAzureBlobStorage(cmd.Context(), binaryBytes, uploadResponse.UploadUrl)
	if err != nil {
		return err
	}

	// Step 3: Inform backend about the completed upload
	developerNote := args[1]
	if err != nil {
		return err
	}

	queryString := `mutation($continuesToken:String!,$uploadMessage: String!) {
		result: pushSnap(input:{
			continuesToken: $continuesToken,
			uploadMessage : $uploadMessage
		}) {
			snapDeclarationId
			snapRevisionId
			fileName
			snapName
			version
			revision
			architecture
		}
	}`

	client, req := gql.PrepareClientAndRequest(cmd.Context(), queryString)
	req.Var("continuesToken", uploadResponse.ContinuesToken)
	req.Var("uploadMessage", developerNote)

	var pushSnapResult struct {
		Info gql.PushSnapOutput `json:"result"`
	}
	err = client.Run(cmd.Context(), req, &pushSnapResult)
	if err != nil {
		return err
	}

	res := SnapPushResult{
		UploadMessage: developerNote,
		FileSize:      snapFileSize,
		Rating:        "",
		Result:        pushSnapResult.Info,
	}

	// Step 3: Set the rating if it was provided
	if ratingId != "" {
		_, err = backend.UpdateSnapRevisionStatus(cmd.Context(), &backend.SnapRevisionUpdateStatusInput{
			Id:           pushSnapResult.Info.SnapRevisionId,
			SnapStatusId: ratingId,
			Description:  descriptionInput,
		})

		if err != nil {
			return err
		}
		res.Rating = ratingInput
	}

	// Print output
	err = format.PrintFormattedOutput(cmd, res, customSnapPushFormatter)
	if err != nil {
		return err
	}

	return nil
}

func customSnapPushFormatter(res SnapPushResult) (string, error) {
	msg := fmt.Sprintf("pushed snap '%s' version '%s' as revision '%d' (%d bytes)", res.Result.SnapName,
		res.Result.Version, res.Result.Revision, res.FileSize)
	if res.Rating != "" {
		msg += fmt.Sprintf(" with rating '%s'", res.Rating)
	}
	return msg, nil
}
