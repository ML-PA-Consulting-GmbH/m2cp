package device

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:   "modify [deviceSerial or deviceName]",
	Short: "Modify device parameters",
	Args:  cobra.ExactArgs(1),
	RunE:  runModifyCmd,
}

type DeviceModifyResult struct {
	Message string `json:"message"`
}

func init() {
	DeviceCmd.AddCommand(modifyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	modifyCmd.Flags().StringP("name", "n", "", "set new device name (must be unique for this tenant)")
	modifyCmd.Flags().StringP("description", "d", "", "set new device description")
	modifyCmd.Flags().String("device-activated", "", "Set whether the device is enabled (true|false). A disabled device is denied communication with the server.")
	modifyCmd.Flags().String("updates-activated", "", "Set whether the device accepts updates (true|false). Disabling halts all updates from the deployment group until re-enabled, independent of --device-activated.")

}

// parseBoolFlag reads a tri-state true|false flag: returns nil if the flag was
// not provided (leave the field unchanged), or a pointer to the boolean value.
func parseBoolFlag(cmd *cobra.Command, name string) (*bool, error) {
	if !cmd.Flags().Changed(name) {
		return nil, nil
	}
	raw, err := cmd.Flags().GetString(name)
	if err != nil {
		return nil, err
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid value %q for --%s: expected true or false", raw, name)
	}
	return tools.BoolPtr(value), nil
}

func runModifyCmd(cmd *cobra.Command, args []string) error {
	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	var nameToUpdate *string
	var descriptionToUpdate *string
	var setToNull []string

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if name == "" {
			setToNull = append(setToNull, gql.DeviceModifyFieldName)
		} else {
			nameToUpdate = tools.StrPtr(name)
		}
	}

	if cmd.Flags().Changed("description") {
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		if description == "" {
			setToNull = append(setToNull, gql.DeviceModifyFieldDescription)
		} else {
			descriptionToUpdate = tools.StrPtr(description)
		}
	}

	active, err := parseBoolFlag(cmd, "device-activated")
	if err != nil {
		return err
	}

	updatesActive, err := parseBoolFlag(cmd, "updates-activated")
	if err != nil {
		return err
	}

	if nameToUpdate == nil && descriptionToUpdate == nil && active == nil && updatesActive == nil && len(setToNull) == 0 {
		return fmt.Errorf("nothing to modify: provide at least one of --name, --description, --device-activated, --updates-activated")
	}

	modifyResult, err := backend.ModifyDevice(cmd.Context(), deviceId, nameToUpdate, descriptionToUpdate, active, updatesActive)
	if err != nil {
		return err
	}

	modified := modifyResult.UpdateDevices[0]
	msg := DeviceModifyResult{
		Message: fmt.Sprintf("Device '%s': deviceActivated=%v, updatesActivated=%v, name='%s', description='%s'",
			modified.SerialNumber,
			modified.IsDeviceActivated,
			modified.IsUpdateActivated,
			tools.MaybeStringToString(modified.DeviceName, "n/a"),
			tools.MaybeStringToString(modified.Description, "n/a"),
		),
	}

	return format.PrintFormattedOutput(cmd, msg, customModifyFormatter)
}

func customModifyFormatter(res DeviceModifyResult) (string, error) {
	list := format.NewList()
	list.Add("Message", res.Message)

	return list.String(), nil
}
