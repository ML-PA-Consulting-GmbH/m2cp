package dgroup

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"m2cpcli/tools/console"
	"time"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [name or Id]",
	Short: "Get information on a Deployment Group",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfoCmd,
}

func init() {
	FleetCmd.AddCommand(infoCmd)
	DGroupCmd.AddCommand(infoCmd)
}

func runInfoCmd(cmd *cobra.Command, args []string) error {
	dgroup, err := backend.GetDeploymentGroupByNameOrId(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, dgroup, customDGroupInfoFormatter)
}

func customDGroupInfoFormatter(dgroup *structs.DeploymentGroup) (string, error) {
	outputStr := ""

	// Generate the fleet tree
	dgroupTree := format.NewTree(console.Colorize(console.Green, "Deployment Group"))
	dgroupTree.AddLeaf("Id: " + dgroup.Id)
	dgroupTree.AddLeaf("Name: " + dgroup.Name)
	dgroupTree.AddLeaf("Auto Update: " + dgroup.AutoUpdate)
	dgroupTree.AddLeaf("Description: " + tools.MaybeStringToString(dgroup.Description, "n/a"))

	modelNode := dgroupTree.NewChild("Model")
	modelNode.AddLeaf("Id: " + dgroup.ModelRevision.Id)
	modelNode.AddLeaf("Type: " + dgroup.ModelRevision.Type)
	modelNode.AddLeaf("Name: " + dgroup.ModelRevision.Name)
	modelNode.AddLeaf("Revision: " + fmt.Sprintf("%d", dgroup.ModelRevision.Revision))
	modelNode.AddLeaf("Architecture: " + dgroup.ModelRevision.Architecture)

	dgroupTree.AddLeaf("Owner: " + dgroup.Owner.Name + " (" + dgroup.Owner.Email + ")")

	if dgroup.CoOwners == nil || len(dgroup.CoOwners) == 0 {
		dgroupTree.AddLeaf("Co-Owners: None")
	} else {
		coAdminsNode := dgroupTree.NewChild("Co-Owners")
		for _, admin := range dgroup.CoOwners {
			coAdminsNode.AddLeaf(admin.Name + " (" + admin.Email + ")")
		}
	}

	dgroupTree.AddLeaf(fmt.Sprintf("Pending Actions: %d", dgroup.PendingActionsTotal))

	outputStr += dgroupTree.String()

	// Devices

	outputStr += "\n"

	outputStr += console.Colorize(console.Green, "Devices") + "\n"
	if len(dgroup.Devices) == 0 {
		outputStr += "None\n"
	} else {
		deviceTable := format.NewTable(map[string]string{
			"serial":      "Serial",
			"name":        "Name",
			"store":       "Latest Store",
			"messaging":   "Latest Messaging",
			"uptime":      "Latest Uptime",
			"pending":     "Pending Actions",
			"description": "Description",
		})

		for _, device := range dgroup.Devices {

			deviceTable.AddRow(map[string]string{
				"serial":      device.DeviceSerial,
				"name":        tools.MaybeStringToString(device.DeviceName, "n/a"),
				"store":       tools.MaybeTimeToString(device.LastAppstoreActivity, time.RFC3339, "n/a"),
				"messaging":   tools.MaybeTimeToString(device.LastMessagingActivity, time.RFC3339, "n/a"),
				"uptime":      tools.MaybeInt64ToStringFormat(device.LastUptime, "n/a", tools.FormatSeconds),
				"pending":     fmt.Sprintf("%d", len(device.DevicePendingActions)),
				"description": tools.ShortenRight(tools.MaybeStringToString(device.Description, "n/a"), 60),
			})
		}
		outputStr += deviceTable.Sort("name").StringSelect([]string{"serial", "name", "store", "messaging", "uptime", "pending", "description"})
	}

	outputStr += "\n" + console.Colorize(console.Green, "App Revisions") + "\n"
	if len(dgroup.AppRevisions) == 0 {
		outputStr += "None\n"
	} else {
		appTable := format.NewTable(map[string]string{
			"core":        "Core",
			"name":        "Name",
			"version":     "Version",
			"rating":      "Rating",
			"installed":   "Installed On",
			"description": "Upload Message/Description",
		})

		for _, app := range dgroup.AppRevisions {
			var description string
			if app.AppDescription != "" {
				description = app.AppDescription
			} else {
				description = app.AppRevisionUploadMessage
			}

			appTable.AddRow(map[string]string{
				"core":        tools.MaybeBoolToString(app.IsSystemApp, "yes", "-", "n/a"),
				"name":        app.AppName,
				"version":     fmt.Sprintf("%s (%d)", app.AppVersion, app.AppRevision),
				"rating":      app.AppRating,
				"installed":   tools.MaybeIntToString(app.InstalledCount, "n/a"),
				"description": tools.ShortenRight(description, 60),
			})
		}
		outputStr += appTable.Sort("name").StringSelect([]string{"core", "name", "version", "rating", "installed", "description"})
	}

	return outputStr, nil
}
