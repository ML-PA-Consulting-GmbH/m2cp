package snapdseed

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/store/system"
)

var snapdSeedCmd = &cobra.Command{
	Use:   "snapd-seed",
	Short: "Commands to work with snapd seed generation",
}

func init() {
	system.SystemCmd.AddCommand(snapdSeedCmd)
}
