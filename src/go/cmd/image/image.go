package image

import (
	"m2cpcli/cmd"

	"github.com/spf13/cobra"
)

// ImageCmd is the parent of all image-inspection subcommands. It does
// not have a behaviour of its own; users reach it via subcommands like
// `m2cp image analyze <image>`.
var ImageCmd = &cobra.Command{
	Use:   "image",
	Short: "Inspect edge-OS image files",
}

func init() {
	cmd.RootCmd.AddCommand(ImageCmd)
}
