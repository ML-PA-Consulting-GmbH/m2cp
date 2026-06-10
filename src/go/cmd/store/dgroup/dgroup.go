package dgroup

import (
	"m2cpcli/cmd/store"

	"github.com/spf13/cobra"
)

// FleetCmd is hidden and exists for backward compatibility
var FleetCmd = &cobra.Command{
	Use:     "fleet",
	Aliases: []string{"f"},
	Short:   "Manage Deployment Groups",
	Hidden:  true,
}

var DGroupCmd = &cobra.Command{
	Use:     "dgroup",
	Aliases: []string{"d"},
	Short:   "Manage Deployment Groups",
}

func init() {
	store.StoreCmd.AddCommand(FleetCmd)
	store.StoreCmd.AddCommand(DGroupCmd)
}
