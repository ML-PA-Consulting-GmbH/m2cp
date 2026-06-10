package coap

import (
	"github.com/spf13/cobra"
)

var requestCmd = &cobra.Command{
	Use:   "request <deviceId|deviceSerial|deviceName> <method> <path+query>",
	Short: "Call a CoAP endpoint on a real time device and return the result",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunCoapArgs(cmd, args)
	},
}

func init() {
	CoapCmd.AddCommand(requestCmd)
}
