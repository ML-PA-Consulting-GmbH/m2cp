package snap

import (
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
)

var listCmd = &cobra.Command{
	Use:   "list [dgroupId | dgroupName]",
	Short: "List the admin and co-administrators of a Deployment Group",
	Args:  cobra.ExactArgs(1),
	RunE:  runListCmd,
}

func init() {
	fleetAdminCmd.AddCommand(listCmd)
}

func runListCmd(cmd *cobra.Command, args []string) error {
	fleetId, err := gql.FleetIdByNameOrId(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	fleet, err := gql.FleetById(cmd.Context(), fleetId, []gql.FleetInfo{gql.FleetAdmin, gql.FleetCoAdmins})
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, fleet, customFleetAdminListFormatter)
}

func customFleetAdminListFormatter(fleet *gql.Fleet) (string, error) {
	table := format.NewTable(map[string]string{
		"role":  "Role",
		"email": "Email",
		"name":  "Name",
	})

	table.AddRow(map[string]string{
		"role":  "Admin",
		"email": fleet.Admin.Email,
		"name":  fleet.Admin.DisplayName,
	})

	for _, coAdmin := range fleet.CoAdmins {
		table.AddRow(map[string]string{
			"role":  "Co-admin",
			"email": coAdmin.Email,
			"name":  coAdmin.DisplayName,
		})
	}

	return table.Sorts([]string{"role", "name"}).StringSelect([]string{"role", "email", "name"}), nil
}
