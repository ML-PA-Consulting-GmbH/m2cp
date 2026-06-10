package ssh

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/device"
)

var sshCmd = &cobra.Command{
	Use:     "ssh",
	Aliases: []string{"s"},
	Short:   "Manage remote connection to device",
}

func init() {
	device.DeviceCmd.AddCommand(sshCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
