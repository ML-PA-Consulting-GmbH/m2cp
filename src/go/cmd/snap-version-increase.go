//go:build !windows

package cmd

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/snap"
	"m2cpcli/tools"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var snapVersionIncreaseCmd = &cobra.Command{
	Use:   "snap-version-inc",
	Short: "Increment or set version of local snap file",
	Long: `Increment or set version of local snap file. It will create a copy of the snap file with the smallest block
of the version number incremented by one - or set a specific version number if given. The version format must match
[0-9]+(\.[0-9]+)*(-[a-zA-Z0-9]+)?, for examples: 123, 123.456, 12.34.56, 1.2.3-beta.`,
	Args: cobra.ExactArgs(1), // TODO: maybe validate with RegExp?
	RunE: runSnapVersionIncreaseCmd,
}

func init() {
	RootCmd.AddCommand(snapVersionIncreaseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	snapVersionIncreaseCmd.Flags().StringP("version", "v", "", "force a specific version")
}

type SnapVersionIncreaseResult struct {
	Warning         string `json:"warning"`
	PreviousVersion string `json:"previousVersion"`
	UpstreamVersion string `json:"upstreamVersion"`
	NewSnapVersion  string `json:"newSnapVersion"`
	SnapFilepath    string `json:"snapFilepath"`
}

func runSnapVersionIncreaseCmd(cmd *cobra.Command, args []string) error {
	snapFile := args[0]
	absoluteSnapFilePath, err := tools.Abspath(snapFile)
	if err != nil {
		return err
	}

	tempDir := filepath.Join(os.TempDir(), uuid.New().String())
	err = os.Mkdir(tempDir, 0700)
	if err != nil {
		return err
	}

	err = exec.Command("unsquashfs", "-n", "-f", "-d", tempDir, absoluteSnapFilePath).Run()
	if err != nil {
		return fmt.Errorf("failed unsquashing snap: %s", err.Error())
	}

	pathYaml := filepath.Join(tempDir, "meta", "snap.yaml")
	dataYaml, err := os.ReadFile(pathYaml)
	if err != nil {
		return err
	}

	declaration, err := snap.UnmarshalDeclaration(dataYaml)
	if err != nil {
		return err
	}

	var versionStore string
	versionOld := declaration.Version

	if cmd.Flags().Changed("version") {
		declaration.Version, err = cmd.Flags().GetString("version")
		if err != nil {
			return err
		}
		if !tools.IsValidSnapVersion(declaration.Version) {
			return fmt.Errorf("invalid version string")
		}
	} else {
		versionStore, err = getLatestVersion(cmd, declaration.Name, declaration.Architectures[0])
		if err != nil {
			return err
		}

		declaration.Version, err = snap.VersionInc(versionStore)
	}

	if err != nil {
		return err
	}
	dataYaml, err = snap.MarshalDeclaration(declaration)
	if err != nil {
		return err
	}

	err = os.WriteFile(pathYaml, dataYaml, 0644)
	if err != nil {
		return err
	}

	pathOut := filepath.Base(absoluteSnapFilePath)
	if strings.Contains(pathOut, versionOld) {
		pathOut = strings.Replace(pathOut, versionOld, declaration.Version, 1)
	} else {
		pathOut = pathOut[:len(pathOut)-5] + "_" + declaration.Version + "_" + declaration.Architectures[0] + ".snap"
	}
	pathOut = filepath.Join(filepath.Dir(absoluteSnapFilePath), pathOut)
	err = exec.Command("mksquashfs", tempDir, pathOut, "-noappend", "-comp", "xz", "-no-fragments", "-no-progress", "-all-root", "-no-xattrs" /*, "-no-exports"*/).Run()
	if err != nil {
		return err
	}

	res := SnapVersionIncreaseResult{
		"this is an experimental feature. don't use with snaps of type 'os', 'core' or 'base'",
		versionOld,
		versionStore,
		declaration.Version,
		pathOut,
	}

	return format.PrintFormattedOutput(cmd, res, customSnapVersionIncreaseFormatter)
}

func customSnapVersionIncreaseFormatter(res SnapVersionIncreaseResult) (string, error) {
	str := fmt.Sprintf("WARNING: %s\n", res.Warning)
	str += fmt.Sprintf("latest version in store: %s\n", res.UpstreamVersion)
	str += fmt.Sprintf("raised version to: %s\n", res.NewSnapVersion)
	str += fmt.Sprintf("output file: %s\n", res.SnapFilepath)
	return str, nil
}

func getLatestVersion(cmd *cobra.Command, snapName string, architecture string) (string, error) {
	queryString := `query ListSnapRevisionsByName($snapName: String! $snapArch: String!){
snapDeclarations(where: {and: [{snapName: {eq: $snapName}}, {snapDeviceArchitecture: {eq: $snapArch}}]}){
	items{
		id
		snapId
		snapRevisions { revision snapVersion }
	}
	totalCount
}}`

	client, req := gql.PrepareClientAndRequest(cmd.Context(), queryString)
	req.Var("snapName", snapName)
	req.Var("snapArch", architecture)
	var result struct {
		SnapDeclarations gql.SnapDeclarationCollectionSegment `json:"snapDeclarations"`
	}
	err := client.Run(cmd.Context(), req, &result)
	if err != nil {
		return "", err
	}
	if result.SnapDeclarations.TotalCount == 0 {
		return "", fmt.Errorf("could not find a matching snap declaration")
	}
	if result.SnapDeclarations.TotalCount > 1 {
		return "", fmt.Errorf("more than one matching snap declaration")
	}

	snapRevisions := result.SnapDeclarations.Items[0].SnapRevisions
	latestRevision, err := latestRevisionByRevisionNumber(snapRevisions)
	if err != nil {
		return "", err
	}

	return latestRevision.SnapVersion, nil
}

func latestRevisionByRevisionNumber(revisions []gql.SnapRevision) (*gql.SnapRevision, error) {
	if len(revisions) > 0 {
		sort.Slice(revisions[:], func(i, j int) bool {
			return revisions[i].Revision < revisions[j].Revision
		})
		return &revisions[len(revisions)-1], nil
	}
	return nil, fmt.Errorf("input list is empty")
}
