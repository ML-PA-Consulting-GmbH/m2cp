package device

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"time"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:   "ping [deviceSerial or deviceName]",
	Short: "Send ping to a m2cp address and return uptime of the reached node",
	Args:  cobra.ExactArgs(1),
	RunE:  runPingCmd,
}

func init() {
	DeviceCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	//pingCmd.Flags().String("device-hub", "", "development: use alternative device hub")
}

type PingResult struct {
	RoundTrip time.Duration `json:"round-trip-time"`
	Uptime    time.Duration `json:"uptime"`
	TaskId    string        `json:"task-id"`
}

func runPingCmd(cmd *cobra.Command, args []string) error {
	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	startTime := time.Now().UTC()
	pingResult, err := gql.EdgeDevicePing(cmd.Context(), gql.UUID(deviceId))
	duration := time.Since(startTime)
	if err != nil {
		return err
	}

	result := PingResult{
		RoundTrip: duration,
		Uptime:    time.Duration(pingResult.Uptime) * time.Second,
		TaskId:    gql.TaskId,
	}

	return format.PrintFormattedOutput(cmd, result, customPingFormatter)
}

func customPingFormatter(res PingResult) (string, error) {
	list := format.NewList()
	list.Add("round trip time", fmt.Sprintf("%s", res.RoundTrip.Round(time.Millisecond)))
	list.Add("uptime", fmt.Sprintf("%s", res.Uptime.Round(time.Second))) // TODO: more human-readable format?
	list.Add("task ID", res.TaskId)

	return list.String(), nil
}
