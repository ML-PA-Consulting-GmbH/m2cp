package message

import (
	"fmt"
	"m2cpcli/cmd/device"

	"github.com/spf13/cobra"
)

var messageCmd = &cobra.Command{
	Use:     "message",
	Aliases: []string{"n"},
	Short:   "Work with device messages",
}

func init() {
	device.DeviceCmd.AddCommand(messageCmd)
	listCmd.Example = fmt.Sprintf("%s list my-device $(date --iso)", messageCmd.CommandPath()) // command path is only set during init
}
