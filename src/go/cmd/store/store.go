package store

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd"
)

var StoreCmd = &cobra.Command{
	Use:     "store",
	Aliases: []string{"s"},
	Short:   "Manage the snap store",
}

func init() {
	cmd.RootCmd.AddCommand(StoreCmd)
}
