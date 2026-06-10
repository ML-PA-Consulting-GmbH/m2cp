package snap

import (
	"fmt"
	"m2cpcli/backend"
	"strings"

	"github.com/spf13/cobra"
)

var sha3Cmd = &cobra.Command{
	Use:   "sha3 <snapName> <arch> <version>",
	Short: "Get SHA3 hash of a snap in your store",
	Args:  cobra.ExactArgs(3),
	RunE:  runSha3Cmd,
}

func init() {
	SnapCmd.AddCommand(sha3Cmd)
}

func runSha3Cmd(cmd *cobra.Command, args []string) error {
	snapName := args[0]
	snapDeviceArchitecture := args[1]
	snapVersion := args[2]

	arch := backend.Architecture(strings.ToUpper(snapDeviceArchitecture))
	if app, err := backend.GetAppByNameAndArchitectureAndVersion(cmd.Context(), &snapName, &arch, &snapVersion); err != nil {
		return err
	} else if app.Apps != nil && app.Apps.Items != nil && len(app.Apps.Items) == 1 && app.Apps.Items[0].AppRevisions != nil && len(app.Apps.Items[0].AppRevisions) > 0 {
		fmt.Printf("%s\n", app.Apps.Items[0].AppRevisions[0].HashSha3)
		return nil
	}
	return fmt.Errorf("not found")
}
