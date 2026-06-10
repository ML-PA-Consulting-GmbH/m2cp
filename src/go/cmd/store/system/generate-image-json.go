package system

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/helper"
	"m2cpcli/tools"
	"strconv"
)

var generateImageJson = &cobra.Command{
	Use:   "generate-image-json [modelName revision]",
	Short: "Generate empty image json file for a model",
	Args:  cobra.ExactArgs(2),
	RunE:  runGenerateImageJsonCmd,
}

func init() {
	SystemCmd.AddCommand(generateImageJson)
}

func runGenerateImageJsonCmd(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	if args[1] == "latest" {
		return fmt.Errorf("latest as revision not supported, please specify a revision number")
	}
	if !tools.IsValidRevision(args[1]) {
		return fmt.Errorf("invalid revision")
	}
	revision, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	modelAssertion, err := helper.GetModelAssertionByModelNameAndRevision(cmd.Context(), modelName, revision)
	if err != nil {
		return nil
	}

	arch, err := helper.GetAssertionFieldValue(modelAssertion.AssertionBody, "architecture")
	if err != nil {
		return err
	}

	parsedSnaps, err := parseAndValidateModel(modelAssertion.AssertionBody)
	if err != nil {
		return err
	}

	var imageSnaps []ImageSnap
	var hasSnapd bool = false
	for _, snap := range parsedSnaps {
		if snap == "snapd" {
			hasSnapd = true
		}
		imageSnap := ImageSnap{
			SnapName:     snap,
			SnapRevision: -1,
		}
		imageSnaps = append(imageSnaps, imageSnap)
	}

	if !hasSnapd {
		imageSnap := ImageSnap{
			SnapName:     "snapd",
			SnapRevision: -1,
		}
		imageSnaps = append(imageSnaps, imageSnap)
	}

	imageJson := ImageJson{
		ModelName:     modelName,
		Architecture:  arch,
		Snaps:         imageSnaps,
		ModelRevision: modelAssertion.Revision,
	}
	jsonOutput, err := json.MarshalIndent(imageJson, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonOutput))
	return nil
}

func parseAndValidateModel(model string) ([]string, error) {
	isClassic, err := helper.GetAssertionFieldValue(model, "classic")
	if err != nil {
		return nil, err
	}
	if isClassic == "true" {
		return nil, fmt.Errorf("classic models are not supported for an image")
	}

	return parseSnapsOfModelAssertion(model)
}
