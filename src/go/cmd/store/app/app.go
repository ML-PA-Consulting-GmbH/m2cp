package app

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/store"
)

var AppCmd = &cobra.Command{
	Use:     "app",
	Aliases: []string{"a"},
	Short:   "Manage apps in the store",
}

func init() {
	store.StoreCmd.AddCommand(AppCmd)
}
