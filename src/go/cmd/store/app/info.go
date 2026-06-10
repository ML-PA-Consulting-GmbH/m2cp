package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"
	"m2cpcli/tools/console"
	"strconv"
	"strings"
)

type output struct {
	AppInfo      backend.GetAppInfoResponse
	AppRevisions backend.GetAppRevisionsByAppIdResponse
}

var infoCmd = &cobra.Command{
	Use:   "info <appId | appName> <arch>",
	Short: "Get information on an app in your store",
	Args:  validateAppInfoArgs,
	RunE:  runInfoCmd,
}

func init() {
	AppCmd.AddCommand(infoCmd)
}

func validateAppInfoArgs(cmd *cobra.Command, args []string) error {
	if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
		return err
	}

	switch len(args) {
	case 1:
		appId := args[0]
		if !tools.IsValidUuid(appId) {
			return fmt.Errorf("invalid appId '%s'", appId)
		}
	case 2:
		appName := args[0]
		arch := args[1]
		if !tools.IsValidAppName(appName) {
			return fmt.Errorf("invalid appName '%s'", appName)
		}
		if !tools.IsValidDeviceArchitecture(arch) {
			return fmt.Errorf("invalid Architecture '%s'", arch)
		}
	default:
		return fmt.Errorf("too many arguments")
	}

	return nil
}

func runInfoCmd(cmd *cobra.Command, args []string) error {
	appId := args[0]

	if len(args) == 2 {
		appName := args[0]
		arch := strings.ToUpper(args[1])
		appIds, err := backend.GetAppIdsByAppNameAndArchitecture(cmd.Context(), appName, backend.Architecture(arch))
		if err != nil {
			return err
		}

		if len(appIds.Apps.Items) == 0 {
			return fmt.Errorf("no app found with name '%s' and architecture '%s'", appName, arch)
		}

		if len(appIds.Apps.Items) > 1 {
			return fmt.Errorf("multiple apps found with name '%s' and architecture '%s', please specify appId", appName, arch)
		}

		appId = appIds.Apps.Items[0].Id
	}

	appInfo, err := backend.GetAppInfo(cmd.Context(), appId)
	if err != nil {
		return err
	}

	appRevisions, err := backend.GetAppRevisionsByAppId(cmd.Context(), appId, 1000, 0)
	if err != nil {
		return err
	}

	outputObj := output{
		AppInfo:      *appInfo,
		AppRevisions: *appRevisions,
	}

	return format.PrintFormattedOutput(cmd, outputObj, customAppInfoFormatter)
}

func customAppInfoFormatter(outputObj output) (string, error) {
	app := outputObj.AppInfo.App

	outputStr := ""

	appInfoTree := format.NewTree(console.Colorize(console.Green, "App Info"))
	appInfoTree.AddLeaf("Id: " + app.Id)
	appInfoTree.AddLeaf("Name: " + app.AppName)
	appInfoTree.AddLeaf("Type: " + getAppTypeName(app.DeviceTypeId))
	appInfoTree.AddLeaf("Architecture: " + strings.ToLower(string(app.Architecture)))

	tenantNode := appInfoTree.NewChild("Tenant")
	tenantNode.AddLeaf("Id: " + app.TenantId)
	tenantNode.AddLeaf("Alias: " + app.Tenant.Alias)

	if app.AppSnap != nil {
		snapNode := appInfoTree.NewChild("Snap")

		snapNode.AddLeaf("SnapId: " + app.AppSnap.SnapId)
		snapNode.AddLeaf("AssertionId: " + app.AppSnap.AssertionId)
	}
	appInfoTree.AddLeaf("Summary: " + app.Summary)
	appInfoTree.AddLeaf("Description: " + app.Description)
	appInfoTree.AddLeaf("Created: " + app.CreatedAt)

	outputStr += appInfoTree.String()

	outputStr += "\n"
	outputStr += console.Colorize(console.Green, "App Revisions") + "\n"

	table := format.NewTable(map[string]string{
		"revision":    "Revision",
		"version":     "Version",
		"rating":      "Rating",
		"description": "Description",
		"size":        "Size",
		"fleetsCount": "#Fleets",
	})

	items := outputObj.AppRevisions.AppRevisions.Items
	for _, item := range items {

		desc := ""
		if item.Description != nil {
			desc = *item.Description
		}

		table.AddRow(map[string]string{
			"revision":    strconv.Itoa(item.Revision),
			"version":     item.Version,
			"rating":      item.AppStatus.Name,
			"description": tools.ShortenRight(desc, 50),
			"size":        tools.FormatBytes(item.DownloadSize),
			"fleetsCount": fmt.Sprintf("%d", len(item.DeploymentGroupBridgeAppRevisions)),
		})
	}

	outputStr += table.Sort("revision:nasc").StringSelect([]string{"revision", "version", "rating", "description", "size", "fleetsCount"})

	return outputStr, nil
}
