package app

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"
	"strings"

	"github.com/spf13/cobra"
)

type appTypeId string

const (
	snap     appTypeId = "8ca69cde-2b48-4aa0-a0d2-422f63ab3f75"
	firmware appTypeId = "9b1c3b9e-e61b-499e-8c95-ff42c7556f29"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available apps in the store",
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

type listOutput struct {
	items []listOutputItem
}

type listOutputItem struct {
	id           string
	architecture string
	appTypeName  string
	appName      string
	tenantAlias  string
	description  string
	shared       string
}

func init() {
	AppCmd.AddCommand(listCmd)

	listCmd.Flags().StringP("name", "n", "", "filter by name containing the given string")
	listCmd.Flags().StringP("arch", "a", "", "filter by architecture (amd64, arm64)")
	listCmd.Flags().StringP("type", "t", "all", "filter by type (all, snap, firmware)")
}

func getAppListFilter(cmd *cobra.Command) (*backend.AppFilterInput, error) {
	conditions := make([]*backend.AppFilterInput, 0)

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, &backend.AppFilterInput{AppName: &backend.StringOperationFilterInput{Contains: &name}})
	}

	if cmd.Flags().Changed("arch") {
		arch, err := cmd.Flags().GetString("arch")
		if err != nil {
			return nil, err
		}

		if !tools.IsValidDeviceArchitecture(arch) {
			return nil, fmt.Errorf("invalid architecture: %s", arch)
		}

		archType := backend.Architecture(strings.ToUpper(arch))
		conditions = append(conditions, &backend.AppFilterInput{Architecture: &backend.ArchitectureOperationFilterInput{Eq: &archType}})
	}

	if cmd.Flags().Changed("type") {
		appType, err := cmd.Flags().GetString("type")
		if err != nil {
			return nil, err
		}

		snapTypeId := string(appTypeId(snap))
		firmwareTypeId := string(appTypeId(firmware))

		switch strings.ToLower(appType) {
		case "snap":
			conditions = append(conditions, &backend.AppFilterInput{DeviceTypeId: &backend.ComparableGuidOperationFilterInput{Eq: &snapTypeId}})
		case "firmware":
			conditions = append(conditions, &backend.AppFilterInput{DeviceTypeId: &backend.ComparableGuidOperationFilterInput{Eq: &firmwareTypeId}})
		case "all":
			break
		default:
			return nil, fmt.Errorf("invalid type: %s, expected snap, firmware or all", appType)
		}
	}

	filter := backend.AppFilterInput{
		And: conditions,
	}

	return &filter, nil
}

func getAppTypeName(deviceTypeId string) string {
	if deviceTypeId == string(snap) {
		return "snap"
	} else if deviceTypeId == string(firmware) {
		return "firmware"
	} else {
		return "unknown"
	}

}

func runListCmd(cmd *cobra.Command, args []string) error {
	filter, err := getAppListFilter(cmd)
	if err != nil {
		return err
	}

	take := 100
	skip := 0
	hasNextPage := true
	items := make([]listOutputItem, 0)

	for hasNextPage {
		result, err := backend.GetAppList(cmd.Context(), filter, &take, &skip)

		if err != nil {
			return err
		}

		if result.Apps != nil && result.Apps.Items != nil && len(result.Apps.Items) > 0 {
			for _, item := range result.Apps.Items {
				items = append(items, listOutputItem{
					id:           item.Id,
					architecture: strings.ToLower(string(item.Architecture)),
					appTypeName:  getAppTypeName(item.DeviceTypeId),
					appName:      item.AppName,
					tenantAlias:  item.Tenant.Alias,
					description:  item.Description,
				})
			}
		} else {
			break
		}

		hasNextPage = result.Apps.PageInfo.HasNextPage
		skip += take
	}

	appIds := make([]string, len(items))
	for i, item := range items {
		appIds[i] = item.id
	}

	sharedStatus, sharedSupported, err := backend.GetAppsGlobalShareStatus(cmd.Context(), appIds)
	if err != nil {
		return err
	}
	for i := range items {
		switch {
		case !sharedSupported:
			items[i].shared = "n/a"
		case sharedStatus[items[i].id]:
			items[i].shared = "true"
		default:
			items[i].shared = "false"
		}
	}

	output := listOutput{
		items: items,
	}
	return format.PrintFormattedOutput(cmd, output, customAppListFormatter)
}

func customAppListFormatter(result listOutput) (string, error) {
	items := result.items

	table := format.NewTable(map[string]string{
		"id":           "Id",
		"architecture": "Architecture",
		"appTypeName":  "App Type",
		"appName":      "App Name",
		"tenantAlias":  "Tenant",
		"description":  "Description",
		"shared":       "Shared",
	})

	for _, item := range items {
		table.AddRow(map[string]string{
			"id":           item.id,
			"architecture": item.architecture,
			"appTypeName":  item.appTypeName,
			"appName":      item.appName,
			"tenantAlias":  item.tenantAlias,
			"description":  strings.TrimSpace(tools.ShortenRight(item.description, 40)),
			"shared":       item.shared,
		})
	}

	outputStr := table.Sorts([]string{"appTypeName", "appName"}).StringSelect([]string{"id", "architecture", "appTypeName", "appName", "tenantAlias", "description", "shared"})
	return outputStr, nil
}
