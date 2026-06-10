package system

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var generateVirtualDeviceCmd = &cobra.Command{
	Use:   "generate-virtual-device [modelName storeName containerRegistry]",
	Short: "Generate a virtual device",
	Long: `Generate a virtual device. The store Name represents the Name of the device image. The container
registry is the URL of the container registry where the device image is stored`,
	Args: cobra.ExactArgs(3),
	RunE: runGenerateVirtualDeviceCmd,
}

//go:embed template/virtual-device
var virtualDeviceTemplate embed.FS
var virtualDeviceTemplateBase = "template/virtual-device"

// Snaps that are installed by default on a virtual device. If you want to add more snaps by default, just add them here
// Currently, the latest revision is used and there is no way to specify a specific revision
var virtualDeviceSnaps = []string{"snapd", "core20", "m2cp-message-hub", "m2cp-gateway"}

// snapData contains all necessary information about a snap that is needed to generate a virtual device
type snapData struct {
	name        string
	revision    int32
	downloadUrl string
	assertion   string
}

type seedSnap struct {
	Name    string `yaml:"name"`
	Channel string `yaml:"channel"`
	File    string `yaml:"file"`
}

// seed is the structure of the seed.yaml that is read by snapd when device is initialized the first time
type seed struct {
	Snaps []*seedSnap `yaml:"snaps"`
}

func init() {
	SystemCmd.AddCommand(generateVirtualDeviceCmd)
}

func runGenerateVirtualDeviceCmd(cmd *cobra.Command, args []string) error {
	modelName := args[0]
	storeName := args[1]
	containerRegistry := args[2]

	// Check, if directory already exists
	virtualDeviceOutputPath := fmt.Sprintf("virtual-device-store-%s", storeName)
	if _, err := os.Stat(virtualDeviceOutputPath); err == nil {
		return fmt.Errorf("directory '%s' already exists", virtualDeviceOutputPath)
	}

	cmd.Println("Fetching model")
	modelAssertion, err := helper.GetModelAssertionByModelNameAndLatestRevision(cmd.Context(), modelName)
	if err != nil {
		return err
	}

	cmd.Println("Validating model")
	err = validateModel(modelAssertion)
	if err != nil {
		return err
	}

	cmd.Println("Fetching snap information")
	snapDataList, err := getSnapData(cmd)
	if err != nil {
		return err
	}

	cmd.Println("Fetching store assertions")
	storeAssertions, err := gql.StoreAssertions(cmd.Context())
	if err != nil {
		return err
	}

	cmd.Println("Preparing structure for virtual device")
	err = prepareFolderStructure(virtualDeviceOutputPath)
	if err != nil {
		return err
	}

	cmd.Println("Updating build script")
	err = replaceTokensInFile(fmt.Sprintf("%s/build.sh", virtualDeviceOutputPath), storeName, containerRegistry)
	if err != nil {
		return err
	}
	cmd.Println("Updating push script")
	err = replaceTokensInFile(fmt.Sprintf("%s/push.sh", virtualDeviceOutputPath), storeName, containerRegistry)
	if err != nil {
		return err
	}
	cmd.Println("Updating run script")
	err = replaceTokensInFile(fmt.Sprintf("%s/run.sh", virtualDeviceOutputPath), storeName, containerRegistry)
	if err != nil {
		return err
	}

	cmd.Println("Preparing seed file")
	seed := prepareSeedFile(snapDataList)
	cmd.Println("Writing seed file")
	err = writeSeedFile(seed, fmt.Sprintf("%s/seed/seed.yaml", virtualDeviceOutputPath))
	if err != nil {
		return err
	}

	cmd.Println("Writing assertions")
	err = writeAssertions(modelAssertion, storeAssertions, snapDataList, fmt.Sprintf("%s/seed/assertions", virtualDeviceOutputPath))
	if err != nil {
		return err
	}

	// Download snaps
	for _, snap := range snapDataList {
		cmd.Println("Downloading snap", snap.name)
		_, err = helper.DownloadSnap(
			snap.downloadUrl, fmt.Sprintf("%s/seed/snaps/%s_%d.snap", virtualDeviceOutputPath, snap.name, snap.revision), "", true)
	}

	cmd.Println(fmt.Sprintf("Done. Virtual device is ready at '%s'", virtualDeviceOutputPath))
	cmd.Println(fmt.Sprintf("WARNING: You still need to put a snapd.deb into the %s root directory!", virtualDeviceOutputPath))
	return nil
}

func replaceTokensInFile(filePath string, storeName string, containerRegistry string) error {
	dict := map[string]string{
		"{{ container-registry }}": containerRegistry,
		"{{ store-name }}":         storeName,
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	for key, value := range dict {
		content = []byte(strings.ReplaceAll(string(content), key, value))
	}

	err = os.WriteFile(filePath, content, 0755)
	if err != nil {
		return err
	}
	return nil
}

func writeAssertions(modelAssertion *gql.Assertion, storeAssertions []*gql.AssertionPlain, snapData []snapData,
	outputPath string) error {
	// Write model assertion
	err := os.WriteFile(fmt.Sprintf("%s/model", outputPath), []byte(modelAssertion.AssertionBody), 0644)
	if err != nil {
		return fmt.Errorf("failed to write model assertion: %w", err)
	}

	// Write store assertions
	formattedStoreAssertions := helper.FormatAssertions(storeAssertions)
	err = os.WriteFile(fmt.Sprintf("%s/store", outputPath), []byte(formattedStoreAssertions), 0644)
	if err != nil {
		return fmt.Errorf("failed to write store assertions: %w", err)
	}

	// Write snap assertions
	for _, snap := range snapData {
		err = os.WriteFile(fmt.Sprintf("%s/%s_%d.assert", outputPath, snap.name, snap.revision),
			[]byte(snap.assertion), 0644)
		if err != nil {
			return fmt.Errorf("failed to write snap assertion: %w", err)
		}
	}

	return nil
}

func writeSeedFile(seed *seed, outputPath string) error {
	// Write seed.yaml
	seedYaml, err := yaml.Marshal(seed)
	if err != nil {
		return fmt.Errorf("failed to marshal seed: %w", err)
	}

	err = os.WriteFile(outputPath, seedYaml, 0644)
	if err != nil {
		return fmt.Errorf("failed to write seed.yaml: %w", err)
	}
	return nil
}

func prepareFolderStructure(outputPath string) error {
	// Move all files from virtualDeviceTemplate into the new directory
	err := applyTemplate(outputPath)
	if err != nil {
		return err
	}

	// Create seed directory
	err = os.Mkdir(fmt.Sprintf("%s/seed", outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create seed directory: %w", err)
	}

	// Create snaps directory
	err = os.Mkdir(fmt.Sprintf("%s/seed/snaps", outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create snaps directory: %w", err)
	}

	// Create model directory
	err = os.Mkdir(fmt.Sprintf("%s/seed/assertions", outputPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create assertions directory: %w", err)
	}

	return nil
}

func applyTemplate(destination string) error {
	err := fs.WalkDir(virtualDeviceTemplate, virtualDeviceTemplateBase, func(filePath string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(virtualDeviceTemplateBase, filePath)
		if err != nil {
			return err
		}

		if dir.IsDir() {
			err := os.Mkdir(filepath.Join(destination, relativePath), os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			err := copyFile(filepath.Join(destination, relativePath), virtualDeviceTemplate, filePath)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func copyFile(destination string, fs embed.FS, sourcePath string) error {
	sourceFile, err := fs.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	if strings.HasSuffix(destination, ".sh") {
		err = os.Chmod(destination, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func prepareSeedFile(snapDataList []snapData) *seed {
	var seed seed
	for _, snapData := range snapDataList {
		var snap seedSnap
		snap.Name = snapData.name
		snap.Channel = "stable"
		snap.File = fmt.Sprintf("%s_%d.snap", snapData.name, snapData.revision)
		seed.Snaps = append(seed.Snaps, &snap)
	}
	return &seed
}

func getSnapData(cmd *cobra.Command) ([]snapData, error) {
	var snapDataList []snapData
	for _, snapName := range virtualDeviceSnaps {
		cmd.Println("\t", snapName)

		var snapData snapData
		// Get snap
		snap, err := helper.SnapByNameAndArchitecture(cmd.Context(), snapName, "amd64")
		if err != nil {
			return nil, fmt.Errorf("failed to get snap-declaration %s: %w", snapName, err)
		}
		revision, err := helper.FindRequestedRevision("latest", snap.SnapRevisions)
		if err != nil {
			return nil, fmt.Errorf("failed to find revision for snap %s: %w", snapName, err)
		}
		snapData.name = snapName
		snapData.revision = revision.Revision
		cmd.Println("\t\tusing revision ", revision.Revision)

		// Get declaration and revision assertion
		assertions, err := gql.SnapAssertionsBySnapDatabaseIdAndRevision(cmd.Context(), snap.Id, revision.Revision)
		if err != nil {
			return nil, fmt.Errorf("failed to get assertion for snap %s: %w", snapName, err)
		}
		snapData.assertion = helper.FormatAssertions(assertions)

		// Get download url
		downloadUrl, err := gql.SnapDownloadUrlBySnapRevisionId(cmd.Context(), revision.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get download url for snap %s: %w", snapName, err)
		}
		snapData.downloadUrl = downloadUrl

		snapDataList = append(snapDataList, snapData)
	}

	return snapDataList, nil
}

func validateModel(modelAssertion *gql.Assertion) error {
	arch, err := helper.GetAssertionFieldValue(modelAssertion.AssertionBody, "architecture")
	if err != nil {
		return err
	}
	if arch != "amd64" {
		return fmt.Errorf("only 'amd64' architecture is supported for a virtual-device")
	}
	classicValue, err := helper.GetAssertionFieldValue(modelAssertion.AssertionBody, "classic")
	if err != nil {
		return err
	}
	if classicValue != "true" {
		return fmt.Errorf("model has to be classic")
	}
	return nil
}
