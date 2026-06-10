package rate

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/store/snap"
)

var rateCmd = &cobra.Command{
	Use:     "rate",
	Aliases: []string{"r"},
	Short:   "Manage snap quality ratings in the store",
}

func init() {
	snap.SnapCmd.AddCommand(rateCmd)
}
