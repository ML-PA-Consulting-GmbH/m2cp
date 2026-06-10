package system

import (
	"encoding/json"
	"fmt"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"os"

	"github.com/spf13/cobra"
)

var getImageDataCmd = &cobra.Command{
	Use:   "get-image-data [modelJsonPath outputPath] ",
	Short: "Gets image data for a model",
	Args:  validateGetImageDataArgs,
	RunE:  runGetImageDataCmd,
}

func init() {
	SystemCmd.AddCommand(getImageDataCmd)
	getImageDataCmd.Flags().BoolP("remove", "r", false, "Remove the image data if it already exists")

}

func validateGetImageDataArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 arguments, got %d", len(args))
	}

	// Check that modelJsonPath exists and is a file
	modelJsonPath := args[0]
	stat, err := os.Stat(modelJsonPath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fmt.Errorf("model json path is not a file")
	}

	// Check that output path exists and is a directory
	outputPath := args[1]
	stat, err = os.Stat(outputPath)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return fmt.Errorf("output path is not a directory")
	}
	return nil
}

func runGetImageDataCmd(cmd *cobra.Command, args []string) error {
	removeFlag, _ := cmd.Flags().GetBool("remove")
	if err := checkSnapsAssertionsDirectoriesExist(args[1]); err != nil {
		if removeFlag {
			cmd.Println(fmt.Sprintf("Removing existing data directories"))
			snapsPath := fmt.Sprintf("%s/snaps", args[1])
			assertionsPath := fmt.Sprintf("%s/assertions", args[1])
			err := os.RemoveAll(snapsPath)
			if err != nil {
				return err
			}
			err = os.RemoveAll(assertionsPath)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	cmd.Println("Reading image json")
	imageJsonPath := args[0]
	imageJson, err := readImageJson(imageJsonPath)
	if err != nil {
		return err
	}

	cmd.Println("Fetching model")
	modelAssertion, err := helper.GetModelAssertionByModelNameAndRevision(cmd.Context(), imageJson.ModelName, imageJson.ModelRevision)
	if err != nil {
		return err
	}

	cmd.Println("Validating model and image json")
	if err = validateModelAndJson(modelAssertion.AssertionBody, *imageJson); err != nil {
		return err
	}

	cmd.Println("Fetching revisions from backend")
	imageJson, err = getRevisionAndAssertionOfImageJsonSnaps(cmd, imageJson)
	if err != nil {
		return err
	}

	cmd.Println("Fetching store assertions")
	storeAssertions, err := gql.StoreAssertions(cmd.Context())
	if err != nil {
		return err
	}

	cmd.Println("Preparing structure")
	outputPath := args[1]
	err = prepareStructure(outputPath)
	if err != nil {
		return err
	}

	cmd.Println("Downloading assertions")
	err = downloadAssertions(helper.FormatAssertions(storeAssertions), modelAssertion.AssertionBody,
		*imageJson, fmt.Sprintf("%s/assertions", outputPath))
	if err != nil {
		return err
	}

	cmd.Println("Downloading snaps")
	err = downloadSnaps(cmd, *imageJson, fmt.Sprintf("%s/snaps", outputPath))
	if err != nil {
		return err
	}

	cmd.Println("Done")
	return nil
}

func checkSnapsAssertionsDirectoriesExist(path string) error {
	// Check that snaps and assertions directory do not exist
	snapsPath := fmt.Sprintf("%s/snaps", path)
	_, err := os.Stat(snapsPath)
	if err == nil {
		return fmt.Errorf("snaps directory already exists")
	}
	assertionsPath := fmt.Sprintf("%s/assertions", path)
	_, err = os.Stat(assertionsPath)
	if err == nil {
		return fmt.Errorf("assertions directory already exists")
	}
	return nil
}

func downloadAssertions(storeAssertions string, modelAssertion string, imageJson ImageJson, outputPath string) error {
	err := os.WriteFile(fmt.Sprintf("%s/model.assert", outputPath), []byte(modelAssertion), 0644)
	if err != nil {
		return err
	}
	for _, snap := range imageJson.Snaps {
		err = os.WriteFile(fmt.Sprintf("%s/%s_%d.assert", outputPath, snap.SnapName, snap.SnapRevision),
			[]byte(snap.Assertion), 0644)
		if err != nil {
			return err
		}
	}
	err = os.WriteFile(fmt.Sprintf("%s/store.assert", outputPath), []byte(storeAssertions), 0644)
	return err
}

func downloadSnaps(cmd *cobra.Command, imageJson ImageJson, outputPath string) error {
	for _, snap := range imageJson.Snaps {
		cmd.Println(fmt.Sprintf("Downloading snap %s with revision %d", snap.SnapName, snap.SnapRevision))
		_, err := helper.DownloadSnap(snap.DownloadUrl,
			fmt.Sprintf("%s/%s_%d.snap", outputPath, snap.SnapName, snap.SnapRevision),
			"", true)
		if err != nil {
			return err
		}
	}
	return nil
}

func prepareStructure(outputPath string) error {
	// Create directories
	err := os.MkdirAll(fmt.Sprintf("%s/snaps", outputPath), 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fmt.Sprintf("%s/assertions", outputPath), 0755)
	if err != nil {
		return err
	}
	return nil
}

func parseSnapsOfModelAssertion(model string) ([]string, error) {
	// Get values of all fields with "name"
	snaps, err := helper.GetAssertionFieldValues(model, "name")
	if err != nil {
		return nil, err
	}
	return snaps, nil
}

func getRevisionAndAssertionOfImageJsonSnaps(cmd *cobra.Command, imageJson *ImageJson) (*ImageJson, error) {
	for i, snap := range imageJson.Snaps {
		cmd.Println(fmt.Sprintf("\t%s", snap.SnapName))
		declaration, err := helper.SnapByNameAndArchitecture(cmd.Context(), snap.SnapName, imageJson.Architecture)
		if err != nil {
			return nil, err
		}
		revision, err := helper.FindRequestedRevision(fmt.Sprintf("%d", snap.SnapRevision), declaration.SnapRevisions)
		if err != nil {
			return nil, err
		}
		assertions, err := gql.SnapAssertionsBySnapDatabaseIdAndRevision(cmd.Context(), declaration.Id, revision.Revision)
		if err != nil {
			return nil, err
		}
		imageJson.Snaps[i].Assertion = helper.FormatAssertions(assertions)

		downloadUrl, err := gql.SnapDownloadUrlBySnapRevisionId(cmd.Context(), revision.Id)
		if err != nil {
			return nil, err
		}
		imageJson.Snaps[i].DownloadUrl = downloadUrl
	}

	return imageJson, nil
}

func readImageJson(path string) (*ImageJson, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var imageJson ImageJson
	err = json.Unmarshal(file, &imageJson)
	if err != nil {
		return nil, err
	}
	return &imageJson, nil
}

func validateModelAndJson(modelAssertion string, imageJson ImageJson) error {
	arch, err := helper.GetAssertionFieldValue(modelAssertion, "architecture")
	if err != nil {
		return err
	}
	if arch != imageJson.Architecture {
		return fmt.Errorf("architecture in image json does not match architecture in model assertion")
	}

	snaps, err := parseSnapsOfModelAssertion(modelAssertion)

	// Check, that all snaps in model assertion are present in image json
	for _, snap := range snaps {
		if !imageJsonContainsSnap(imageJson.Snaps, snap) {
			return fmt.Errorf("snap %s of model is not present in image json", snap)
		}
	}

	for _, snap := range imageJson.Snaps {
		if snap.SnapRevision < 1 {
			return fmt.Errorf("revision of snap %s is not valid", snap.SnapName)
		}
	}
	return nil
}

func imageJsonContainsSnap(snaps []ImageSnap, snap string) bool {
	for _, s := range snaps {
		if s.SnapName == snap {
			return true
		}
	}
	return false
}
