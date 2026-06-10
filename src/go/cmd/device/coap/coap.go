package coap

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/cmd/device"
	"m2cpcli/format"
	"strings"

	"github.com/spf13/cobra"
)

// RunE provides silent backwards compatibility with the old
// "m2cp device coap <deviceId> <method> <path+query>" usage.
// When Cobra cannot match the first argument to a known subcommand it
// falls through to this handler. It is intentionally not advertised.

var CoapCmd = &cobra.Command{
	Use:   "coap",
	Short: "CoAP operations on devices",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return cmd.Help()
		}
		return RunCoapArgs(cmd, args)
	},
}

func init() {
	device.DeviceCmd.AddCommand(CoapCmd)
}

// RunCoapArgs is the shared CLI entry point used by both the legacy backwards-
// compatible invocation and the explicit "coap request" subcommand.
// args must be [deviceIdOrSerial, method, path+query].
func RunCoapArgs(cmd *cobra.Command, args []string) error {
	method, path, query, err := parseCoapArgs(args)
	if err != nil {
		return err
	}

	var body []byte

	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	dev, err := backend.GetDeviceInfoById(cmd.Context(), deviceId)
	if err != nil {
		return err
	}

	result, err := device.CoapRequest(cmd.Context(), dev, method, path, query, body)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, *result, CoapFormatter)
}

// parseCoapArgs validates and decomposes [deviceId, method, path+query] args.
func parseCoapArgs(args []string) (method, path, query string, err error) {
	method = strings.ToUpper(args[1])
	if method != "GET" && method != "PUT" && method != "POST" {
		return "", "", "", fmt.Errorf("invalid method '%s'", method)
	}
	tokens := strings.SplitN(args[2], "?", 2)
	path = tokens[0]
	if len(tokens) > 1 {
		query = tokens[1]
	}
	return method, path, query, nil
}

// CoapFormatter is the human-readable output formatter for CoapResultType.
func CoapFormatter(res device.CoapResultType) (string, error) {
	out := ""
	out += fmt.Sprintf("Status Code: %s\n", res.StatusCode)
	out += "Body:\n"
	out += string(res.Response) + "\n"

	return out, nil
}
