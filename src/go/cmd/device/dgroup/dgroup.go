package dgroup

import (
	"m2cpcli/cmd/device"

	"github.com/spf13/cobra"
)

// fleetCmd is hidden and exists for backward compatibility
var fleetCmd = &cobra.Command{
	Use:     "fleet",
	Aliases: []string{"f"},
	Short:   "Manage Deployment Group of the Device",
	Hidden:  true,
}

var dgroupCmd = &cobra.Command{
	Use:     "dgroup",
	Aliases: []string{"d"},
	Short:   "Manage Deployment Group of the Device",
}

func init() {
	device.DeviceCmd.AddCommand(fleetCmd)
	device.DeviceCmd.AddCommand(dgroupCmd)
}
