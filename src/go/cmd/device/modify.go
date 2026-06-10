package device

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"m2cpcli/tools"

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
	modifyCmd.Flags().String("device-activated", "", "Set the activation state of a device (yes|no). Deactivated devices are denied communication with the server.")

}

func runModifyCmd(cmd *cobra.Command, args []string) error {
	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	var nameToUpdate *string
	var descriptionToUpdate *string
	var active *bool
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

	if cmd.Flags().Changed("lifecycle") {
		activationValue, err := cmd.Flags().GetString("lifecycle")
		if err != nil {
			return err
		}
		if activationValue != "yes" && activationValue != "no" {
			return fmt.Errorf("invalid lifecycle state")
		}
		activationBool := activationValue == "yes"
		active = &activationBool
	}

	modifyResult, err := backend.ModifyDevice(cmd.Context(), deviceId, nameToUpdate, descriptionToUpdate, active)
	if err != nil {
		return err
	}

	modified := modifyResult.UpdateDevices[0]
	msg := DeviceModifyResult{
		Message: fmt.Sprintf("Device '%s': active=%v, name='%s', description='%s'",
			modified.SerialNumber,
			modified.IsDeviceActivated,
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
