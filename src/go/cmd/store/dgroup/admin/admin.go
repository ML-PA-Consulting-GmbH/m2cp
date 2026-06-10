package snap

import (
	"m2cpcli/cmd/store/dgroup"

	"github.com/spf13/cobra"
)

var fleetAdminCmd = &cobra.Command{
	Use:     "admin",
	Aliases: []string{"a"},
	Short:   "Manage administrators of the Deployment Group",
}

func init() {
	dgroup.FleetCmd.AddCommand(fleetAdminCmd)
	dgroup.DGroupCmd.AddCommand(fleetAdminCmd)
}
