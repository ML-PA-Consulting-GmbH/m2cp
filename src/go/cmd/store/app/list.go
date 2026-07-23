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
	Items []listOutputItem `json:"items"`
}

type listOutputItem struct {
	Id           string `json:"id"`
	Architecture string `json:"architecture"`
	AppTypeName  string `json:"appType"`
	AppName      string `json:"appName"`
	TenantAlias  string `json:"tenant"`
	Description  string `json:"description"`
	Shared       string `json:"shared"`
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
					Id:           item.Id,
					Architecture: strings.ToLower(string(item.Architecture)),
					AppTypeName:  getAppTypeName(item.DeviceTypeId),
					AppName:      item.AppName,
					TenantAlias:  item.Tenant.Alias,
					Description:  item.Description,
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
		appIds[i] = item.Id
	}

	sharedStatus, sharedSupported, err := backend.GetAppsGlobalShareStatus(cmd.Context(), appIds)
	if err != nil {
		return err
	}
	for i := range items {
		switch {
		case !sharedSupported:
			items[i].Shared = "n/a"
		case sharedStatus[items[i].Id]:
			items[i].Shared = "true"
		default:
			items[i].Shared = "false"
		}
	}

	output := listOutput{
		Items: items,
	}
	return format.PrintFormattedOutput(cmd, output, customAppListFormatter)
}

func customAppListFormatter(result listOutput) (string, error) {
	items := result.Items

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
			"id":           item.Id,
			"architecture": item.Architecture,
			"appTypeName":  item.AppTypeName,
			"appName":      item.AppName,
			"tenantAlias":  item.TenantAlias,
			"description":  strings.TrimSpace(tools.ShortenRight(item.Description, 40)),
			"shared":       item.Shared,
		})
	}

	outputStr := table.Sorts([]string{"appTypeName", "appName"}).StringSelect([]string{"id", "architecture", "appTypeName", "appName", "tenantAlias", "description", "shared"})
	return outputStr, nil
}
