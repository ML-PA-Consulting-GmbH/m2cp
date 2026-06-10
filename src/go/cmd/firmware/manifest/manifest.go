package manifest

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/firmware"
)

var ManifestCmd = &cobra.Command{
	Use:     "manifest",
	Aliases: []string{"m"},
	Short:   "Manage firmware manifests",
}

func init() {
	firmware.FirmwareCmd.AddCommand(ManifestCmd)
}
