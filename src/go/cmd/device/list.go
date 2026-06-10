package device

import (
	"encoding/json"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List existing devices",
	Long:  `Output a list existing devices. Output can be filtered with the flags.`,
	Args:  cobra.ExactArgs(0),
	RunE:  runListCmd,
}

func init() {
	DeviceCmd.AddCommand(listCmd)

	//listCmd.Flags().Bool("activated", false, "only devices with activation status (true, false)")
	listCmd.Flags().String("name", "", "only devices with name containing this string")
	listCmd.Flags().StringP("arch", "a", "", "only devices with architecture (amd64, arm32, arm64)")
	listCmd.Flags().String("after", "", "only devices last app store active >= this date (YYYY-MM-DD, default two weeks before now)")
	listCmd.Flags().String("before", "", "only devices last app store active <= this date (YYYY-MM-DD)")
	listCmd.Flags().String("sort", "id", "sort by (serialNumber, deviceName, lastAppStoreActivity, id)")
	listCmd.Flags().StringP("type", "t", "", fmt.Sprintf("only devices of given type (\"%s\" or \"%s\" or \"ed\" or \"rtd\")",
		structs.DeviceTypeEdgeDevice, structs.DeviceTypeRealTimeDevice))

	listCmd.Flags().Bool("skip-online-status", false, "Skip online status")
	listCmd.Flags().BoolP("online", "o", false, "only online devices")
	listCmd.Flags().Bool("offline", false, "only offline devices")
}

func getDeviceListFilter(cmd *cobra.Command) ([]structs.BackendQueryFilter, error) {
	filters := make([]structs.BackendQueryFilter, 0)

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return nil, err
		}
		filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterName, Pattern: &name})
	}

	if cmd.Flags().Changed("arch") {
		arch, err := cmd.Flags().GetString("arch")
		if err != nil {
			return nil, err
		}
		if !tools.IsValidDeviceArchitecture(arch) {
			return nil, fmt.Errorf("invalid architecture '%s'", arch)
		}
		arch = strings.ToLower(arch)
		filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterArch,
			ExactAny: []string{arch}})
	}

	if cmd.Flags().Changed("type") {
		deviceType, err := cmd.Flags().GetString("type")
		if err != nil {
			return nil, err
		}

		// Try to figure out what the user means
		// "Real-Time Devices" as rtd and like "Edge Device"
		deviceType, _ = tools.CanonicalDeviceTypeFromUserString(deviceType)

		if !tools.IsValidDeviceType(deviceType) {
			return nil, fmt.Errorf("invalid type '%s'", deviceType)
		}

		filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterType,
			ExactAny: []string{deviceType}})
	}

	if cmd.Flags().Changed("after") {
		after, err := cmd.Flags().GetString("after")
		if err != nil {
			return nil, err
		}
		if after != "" {
			parsedAfter, err := time.Parse("2006-01-02", after)
			if err != nil {
				return nil, fmt.Errorf("failed parsing date. '%s' is not in YYYY-MM-DD format: %s", after, err)
			}
			after = parsedAfter.Format(time.RFC3339)
			filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterDeviceLastStoreActivity, After: &parsedAfter})
		}
	} else {
		// Default to two weeks before now
		const fourteenDays = time.Duration(14*24) * time.Hour
		date := time.Now().UTC().Add(-1 * fourteenDays)
		filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterDeviceLastStoreActivity, After: &date})
	}

	if cmd.Flags().Changed("before") {
		before, err := cmd.Flags().GetString("before")
		if err != nil {
			return nil, err
		}
		if before != "" {
			parsedBefore, err := time.Parse("2006-01-02", before)
			if err != nil {
				return nil, fmt.Errorf("failed parsing date. '%s' is not in YYYY-MM-DD format: %s", before, err)
			}
			before = parsedBefore.Format(time.RFC3339)
			filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterDeviceLastStoreActivity, Before: &parsedBefore})
		}
	}

	if cmd.Flags().Changed("sort") {
		sort, err := cmd.Flags().GetString("sort")
		if err != nil {
			return nil, err
		}

		switch sort {
		case "serialNumber":
			filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterOSSerial, Sort: tools.Ptr(structs.BackendQueryFilterSortAsc)})
		case "deviceName":
			filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterName, Sort: tools.Ptr(structs.BackendQueryFilterSortAsc)})
		case "lastAppStoreActivity":
			filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterDeviceLastStoreActivity, Sort: tools.Ptr(structs.BackendQueryFilterSortAsc)})
		case "id":
			filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterId, Sort: tools.Ptr(structs.BackendQueryFilterSortAsc)})
		default:
			return nil, fmt.Errorf("invalid sort key '%s', valid keys are: serialNumber, deviceName, lastAppStoreActivity, id", sort)
		}
	} else {
		filters = append(filters, structs.BackendQueryFilter{Field: structs.BackendQueryFilterId, Sort: tools.Ptr(structs.BackendQueryFilterSortAsc)})
	}

	return filters, nil
}

type listOutput struct {
	Items  []structs.Device             `json:"items"`
	Filter []structs.BackendQueryFilter `json:"filter,omitempty"`
}

func runListCmd(cmd *cobra.Command, args []string) error {

	filter, err := getDeviceListFilter(cmd)
	if err != nil {
		return err
	}

	take := 100
	skip := 0

	hasNextPage := true
	var devices []structs.Device
	for hasNextPage {
		var moreDevices []structs.Device
		moreDevices, hasNextPage, err = backend.GetDeviceList(cmd.Context(), filter, &take, &skip)
		devices = append(devices, moreDevices...)
		if err != nil {
			return err
		}
		skip += take
	}

	output := listOutput{}
	output.Items = []structs.Device{}
	output.Filter = filter

	onlyOnline, _ := cmd.Flags().GetBool("online")
	onlyOffline, _ := cmd.Flags().GetBool("offline")

	for _, device := range devices {
		if device.DeviceType == structs.DeviceTypeRealTimeDevice {
			device.IsOnline = nil // Online status is not available for arm32 devices, so we set it to nil to avoid confusion
		}

		if (onlyOnline && tools.MaybeBoolToBool(device.IsOnline, false)) ||
			(onlyOffline && !tools.MaybeBoolToBool(device.IsOnline, false)) ||
			(!onlyOnline && !onlyOffline) {
			output.Items = append(output.Items, device)
		}
	}

	return format.PrintFormattedOutput(cmd, output, customDeviceListFormatter)
}

func customDeviceListFormatter(res listOutput) (string, error) {
	table := format.NewTable(map[string]string{
		"serial":            "Serial",
		"type":              "Type",
		"arch":              "Arch",
		"name":              "Name",
		"desc":              "Description",
		"storeActivity":     "Store Activity",
		"messagingActivity": "Messaging Activity",
		"online":            "Online",
		"model":             "Model",
		"uplinkMode":        "Uplink Mode",
	})
	for _, device := range res.Items {

		table.AddRow(map[string]string{
			"serial":            device.DeviceSerial,
			"type":              string(device.DeviceType),
			"arch":              device.DeviceArchitecture,
			"name":              tools.MaybeStringToString(device.DeviceName, "n/a"),
			"desc":              strings.TrimSpace(tools.ShortenRight(tools.MaybeStringToString(device.Description, "n/a"), 45)),
			"storeActivity":     tools.MaybeTimeToString(device.LastAppstoreActivity, "2006-01-02", "n/a"),
			"messagingActivity": tools.MaybeTimeToString(device.LastMessagingActivity, "2006-01-02", "n/a"),
			"online":            tools.MaybeBoolToString(device.IsOnline, "true", "false", "n/a"),
			"model":             fmt.Sprintf("%s (%d)", device.DeviceModelRevision.Name, device.DeviceModelRevision.Revision),
			"uplinkMode":        tools.MaybeStringToString(device.UplinkMode, "n/a"),
		})
	}

	outputStr := table.Sort("active:desc").StringSelect([]string{"serial", "type", "arch", "name", "desc", "storeActivity", "messagingActivity", "online", "model", "uplinkMode"})

	jsonBytes, _ := json.Marshal(res.Filter)
	outputStr += fmt.Sprintf("\nFilter: %s\n", string(jsonBytes))

	return outputStr, nil
}
