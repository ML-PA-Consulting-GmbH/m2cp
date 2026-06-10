package cmd

import (
	"fmt"
	"m2cpcli/env"
	"m2cpcli/format"
	"m2cpcli/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version",
	RunE:  runVersionCmd,
}

func init() {
	RootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type VersionResult struct {
	ProgramName string               `json:"program-name"`
	SemVer      *env.SemanticVersion `json:"https://semver.org/"`
	Git         env.GitMetadata      `json:"git"`
}

func runVersionCmd(cmd *cobra.Command, args []string) error {
	var result VersionResult
	var err error
	result.ProgramName = cmd.Parent().Use
	result.SemVer, err = env.NewSemanticVersion(version.Version)
	if err != nil {
		return fmt.Errorf("on parsing version: %s", err)
	}
	result.Git.Commit = version.CommitHash
	return format.PrintFormattedOutput(cmd, result, customVersionFormatter)
}

func customVersionFormatter(res VersionResult) (string, error) {
	str := fmt.Sprintf("%s version %s", res.ProgramName, res.SemVer.String())
	return str, nil
}
