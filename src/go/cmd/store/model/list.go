package model

import (
	"github.com/spf13/cobra"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/tools"
	"strconv"
	"strings"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all models of your store",
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

type ListOutput struct {
	Items []ListOutputItem `json:"edgeDeviceModels"`
}

type ListOutputItem struct {
	Id             string `json:"id"`
	DeviceModelId  string `json:"deviceModelId"`
	DeviceTypeName string `json:"deviceTypeName"`
	Architecture   string `json:"architecture"`
	ModelName      string `json:"modelName"`
	Revision       int    `json:"revision"`
	UploadMessage  string `json:"uploadMessage"`
	Shared         string `json:"shared"`
}

func init() {
	ModelCmd.AddCommand(listCmd)
}

func runListCmd(cmd *cobra.Command, args []string) error {
	take := 100
	skip := 0
	hasNextPage := true
	items := make([]ListOutputItem, 0)

	for hasNextPage {
		result, err := backend.GetDeviceModelRevisionList(cmd.Context(), nil, &take, &skip)

		if err != nil {
			return err
		}

		if result.DeviceModelRevisions != nil && result.DeviceModelRevisions.Items != nil && len(result.DeviceModelRevisions.Items) > 0 {
			for _, item := range result.DeviceModelRevisions.Items {
				revision := 0
				if item.Revision != nil {
					revision = *item.Revision
				}
				items = append(items, ListOutputItem{
					Id:             item.Id,
					DeviceModelId:  item.DeviceModelId,
					Architecture:   strings.ToLower(string(item.DeviceModel.Architecture)),
					DeviceTypeName: getDeviceTypeName(item.DeviceModel.DeviceTypeId),
					ModelName:      item.DeviceModel.ModelName,
					Revision:       revision,
					UploadMessage:  item.UploadMessage,
				})
			}
		} else {
			break
		}

		hasNextPage = result.DeviceModelRevisions.PageInfo.HasNextPage
		skip += take
	}

	deviceModelIdSet := make(map[string]struct{}, len(items))
	for _, item := range items {
		deviceModelIdSet[item.DeviceModelId] = struct{}{}
	}
	deviceModelIds := make([]string, 0, len(deviceModelIdSet))
	for id := range deviceModelIdSet {
		deviceModelIds = append(deviceModelIds, id)
	}

	sharedStatus, sharedSupported, err := backend.GetDeviceModelsGlobalShareStatus(cmd.Context(), deviceModelIds)
	if err != nil {
		return err
	}
	for i := range items {
		switch {
		case !sharedSupported:
			items[i].Shared = "n/a"
		case sharedStatus[items[i].DeviceModelId]:
			items[i].Shared = "true"
		default:
			items[i].Shared = "false"
		}
	}

	output := ListOutput{
		Items: items,
	}

	return format.PrintFormattedOutput(cmd, output, customModelListFormatter)
}

func getDeviceTypeName(deviceTypeId string) string {
	deviceTypeId = strings.ToLower(deviceTypeId)
	if deviceTypeId == "8ca69cde-2b48-4aa0-a0d2-422f63ab3f75" {
		return "m2cp"
	} else if deviceTypeId == "9b1c3b9e-e61b-499e-8c95-ff42c7556f29" {
		return "riot"
	} else {
		return "unknown"
	}
}

func customModelListFormatter(result ListOutput) (string, error) {
	items := result.Items
	table := format.NewTable(map[string]string{
		"id":             "Id",
		"deviceTypeName": "Type",
		"architecture":   "Arch",
		"modelName":      "Name",
		"revision":       "Revision",
		"uploadMessage":  "Upload Message",
		"shared":         "Shared",
	})

	for _, item := range items {
		table.AddRow(map[string]string{
			"id":             item.Id,
			"deviceTypeName": item.DeviceTypeName,
			"architecture":   item.Architecture,
			"modelName":      item.ModelName,
			"revision":       strconv.Itoa(item.Revision),
			"uploadMessage":  strings.TrimSpace(tools.ShortenRight(item.UploadMessage, 60)),
			"shared":         item.Shared,
		})
	}

	outputStr := table.Sorts([]string{"deviceTypeName", "modelName", "revision"}).StringSelect([]string{"id", "deviceTypeName", "modelName", "revision", "architecture", "uploadMessage", "shared"})

	return outputStr, nil
}
