package uptime

import (
	"fmt"
	"m2cp"
	"m2cp/m2cp_new"
	"m2cpcli/backend"
	"m2cpcli/cmd/device/message"
	"m2cpcli/format"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/square/go-jose.v2/json"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze <device> <date>",
	Short: "Fetch and analyze uptime statistics from signal messages for <device> on <date:YYYY-MM-DD>",
	Long:  "Analyzes signal messages to extract uptime information and generate statistics",
	Args:  cobra.ExactArgs(2),
	RunE:  runAnalyzeCmd,
}

func init() {
	uptimeCmd.AddCommand(analyzeCmd)
	analyzeCmd.Flags().String("csv", ";", "generate csv output with given separator")
}

func runAnalyzeCmd(cmd *cobra.Command, args []string) error {
	ctp := m2cp_new.ContextPlusFromContext(cmd.Context())

	parsedDate, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		return fmt.Errorf("failed parsing date. '%s' is not in YYYY-MM-DD format: %s", args[1], err)
	}

	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(ctp, args[0])
	if err != nil {
		return err
	}

	stats, err := analyzeUptimeMessages(ctp, deviceId, parsedDate)
	if err != nil {
		return err
	}

	if csvSeparator := cmd.Flag("csv").Value.String(); csvSeparator != "" {
		return format.PrintFormattedOutput(cmd, stats, func(output []uptimeDetails) (string, error) {
			outputStr := fmt.Sprintf("%s%s%s%s%s%s%s\n", "time", csvSeparator, "uptime", csvSeparator, "requested", csvSeparator, "power-manager")
			for _, item := range output {
				serverStatus := "unknown"
				if item.Server != "" {
					serverStatus = "connected"
				}
				outputStr += fmt.Sprintf("%s%s%d%s%d%s%s\n", item.Time, csvSeparator, item.Running, csvSeparator, item.Remaining, csvSeparator, serverStatus)
			}
			return outputStr, nil
		})
	}
	return format.PrintFormattedOutput(cmd, stats, nil)
}

type heartbeatUptimeDetails struct {
	uptimeDetails `json:"uptime"`
}

type uptimeDetails struct {
	Time      string `json:"time"`
	Running   uint32 `json:"running"`
	Remaining uint32 `json:"remaining"`
	Server    string `json:"server"`
}

func analyzeUptimeMessages(ctp m2cp.ContextPlus, deviceId string, date time.Time) ([]uptimeDetails, error) {
	var uptimeValues []uptimeDetails

	callback := func(topic, origin, datetime string, data map[string]string) {
		if topic != "signal/log" || !strings.HasPrefix(origin, "log.m2cp-coap.") {
			return
		}
		if data["name:string"] != "status" {
			return
		}
		var heartbeat heartbeatUptimeDetails
		if err := json.Unmarshal([]byte(data["content:string"]), &heartbeat); err != nil {
			fmt.Println("WARNING: failed parsing")
			return
		}
		heartbeat.uptimeDetails.Time = datetime
		uptimeValues = append(uptimeValues, heartbeat.uptimeDetails)
	}

	outpath, err := os.MkdirTemp("", "m2cp-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	if _, err = message.ProcessDataMessages(ctp, deviceId, date, outpath, callback); err != nil {
		return nil, err
	}

	return uptimeValues, nil
}
