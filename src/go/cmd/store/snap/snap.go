package snap

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/store"
)

var SnapCmd = &cobra.Command{
	Use:     "snap",
	Aliases: []string{"s"},
	Short:   "Manage snaps in the store",
}

func init() {
	store.StoreCmd.AddCommand(SnapCmd)
}
