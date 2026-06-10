package device

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/f"
	"m2cpcli/tools"
	"m2cpcli/tools/console"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [device-serial or device-name]",
	Short: "Get logs from device",
	Long: `
  The "logs" command retrieves logs from a specified device. 
  You can filter the logs by different criteria, such as log level, 
  origin, topic, task ID, and date range. If no OS Serial Number is provided, 
  the command will use a default device or all devices based on the system configuration.`,
	Example: `  
  # Get logs from a specific device by providing its serial number
  m2cp device logs abc123

  # Get logs with a log level of WARNING or higher from all devices
  m2cp device logs --level WARNING

  # Get logs newer than 1 hour with a specific task ID
  m2cp device logs --newer-than 1h --task mytaskid

  # Search logs by the term "error" within a specific time frame
  m2cp device logs --after 2024-09-10T00:00 --before 2024-09-11T00:00 --search error --limit 10`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogsCmd,
}

func init() {
	DeviceCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringP("level", "", "INFO", "only print logs of given level or above (DEBUG, INFO, WARNING, ERROR)")
	logsCmd.Flags().StringP("origin", "", "", "only print logs stemming from signal message with given origin, * allowed as wildcard")
	logsCmd.Flags().StringP("topic", "", "", "only print logs stemming from signal message with given topic, * allowed as wildcard")
	logsCmd.Flags().StringP("newer-than", "n", "", "only print logs newer than given duration (format: 1m, 1h)")
	logsCmd.Flags().StringP("task", "t", "", "show only logs with a specific task id")
	logsCmd.Flags().StringP("search", "s", "", "search term to filter logs by multiple fields")
	logsCmd.Flags().StringP("after", "", "", "show only logs newer than this UTC date (format: 2006-01-02T15:04)")
	logsCmd.Flags().StringP("before", "", "", "show only logs older than this UTC date (format: 2006-01-02T15:04")
	logsCmd.Flags().IntP("limit", "l", 100, "limit the number of results to the n newest")
}

func runLogsCmd(cmd *cobra.Command, args []string) error {
	filter := backend.EdgeDeviceLogQueryInput{}

	if len(args) > 0 {
		serialOrName := args[0]

		if tools.IsValidUuid(serialOrName) {
			filter.DeviceSerial = &serialOrName
		} else {
			deviceSerial, err := backend.GetDeviceSerialByName(cmd.Context(), serialOrName)
			if err != nil {
				return err
			}
			filter.DeviceSerial = &deviceSerial
		}
	}

	limitStr := cmd.Flag("limit").Value.String()

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return err
		}

		filter.Limit = &limit
	}

	taskId := cmd.Flag("task").Value.String()
	searchString := cmd.Flag("search").Value.String()
	signalType := cmd.Flag("level").Value.String()
	origin := cmd.Flag("origin").Value.String()

	filter.TaskId = &taskId
	filter.SearchString = &searchString
	filter.SignalType = &signalType
	filter.Origin = &origin

	after := cmd.Flag("after").Value.String()
	before := cmd.Flag("before").Value.String()
	newerThan := cmd.Flag("newer-than").Value.String()

	if newerThan != "" && after != "" {
		return fmt.Errorf("newer-than and after cannot be set at the same time")
	}

	if after != "" {
		filter.StartDateTime = &after
	}

	if before != "" {
		filter.EndDateTime = &before
	}

	if newerThan != "" {
		duration, err := time.ParseDuration(newerThan)
		if err != nil {
			return err
		}

		startDateTime := time.Now().Add(-duration).Format(time.RFC3339)
		filter.StartDateTime = &startDateTime
	}

	deviceLogs, err := backend.GetDeviceLogs(cmd.Context(), &filter)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, deviceLogs, deviceLogsFormatter)
}

func deviceLogsFormatter(res *backend.GetDeviceLogsResponse) (string, error) {
	outputStr := ""

	for _, log := range res.Data {
		var formattedTime string
		if log.DateTime != nil {
			if parsedTime, err := time.Parse(time.RFC3339, *log.DateTime); err == nil {
				formattedTime = parsedTime.Format("2006-01-02 15:04:05")
			}
		}

		color := getColorForSignalType(f.MaybeToString(log.SignalType, ""))

		outputStr += fmt.Sprintf("\n──┬─ [%s] ─ %s ", console.Colorize(color, f.MaybeToString(log.SignalType, "n/a")), formattedTime)
		outputStr += fmt.Sprintf("\n  ├─ Device:  %s", f.MaybeToString(log.DeviceSerial, "n/a"))
		outputStr += fmt.Sprintf("\n  ├─ Name:    %s", f.MaybeToString(log.SignalName, "n/a"))
		outputStr += fmt.Sprintf("\n  ├─ Origin:  %s", f.MaybeToString(log.Origin, "n/a"))
		outputStr += fmt.Sprintf("\n  ├─ Topic:   %s", f.MaybeToString(log.Topic, "n/a"))
		outputStr += fmt.Sprintf("\n  ├─ TaskId:  %s", f.MaybeToString(log.TaskId, "n/a"))
		outputStr += fmt.Sprintf("\n  └─ Content: %s\n", f.MaybeToString(log.SignalContent, "n/a"))

	}

	return outputStr, nil
}

func getColorForSignalType(signalType string) console.Color {
	switch signalType {
	case "ERROR":
		return console.Red
	case "WARNING":
		return console.Yellow
	case "INFO":
		return console.Green
	default:
		return console.Reset
	}
}
