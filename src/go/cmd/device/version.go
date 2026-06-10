package device

import (
	"fmt"
	"m2cp/device"
	"m2cp/m2cp_new"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version [deviceSerial or deviceName]",
	Short: "Output or parse a device' OS version string",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runVersionCmd,
}

func init() {
	DeviceCmd.AddCommand(versionCmd)
	versionCmd.Flags().String("parse", "", "parse a given M2CP device OS version")
}

func runVersionCmd(cmd *cobra.Command, args []string) error {
	ctp := m2cp_new.ContextPlusFromContext(cmd.Context())

	if cmd.Flags().Changed("parse") {
		short, err := cmd.Flags().GetString("parse")
		if err != nil {
			return err
		}
		dov, err := device.DecodeOsVersion(short)
		if err != nil {
			return err
		}
		return format.PrintFormattedOutput(cmd, dov, nil)
	} else {
		if len(args) != 1 {
			return fmt.Errorf("deviceSerial or deviceName missing")
		}

		deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
		if err != nil {
			return err
		}

		deviceDetails, err := backend.GetDeviceInfoById(cmd.Context(), deviceId)
		if err != nil {
			return err
		}

		// compile list of installed snaps
		snaps := map[string]uint{}
		for _, item := range deviceDetails.DeviceInstallStates {
			snapName := item.AppName
			snapRevision := item.AppRevision
			snaps[snapName] = uint(snapRevision)
		}

		modelId := deviceDetails.DeviceModelRevision.Id
		model, err := backend.GetDeviceModelByIdWithAssertion(cmd.Context(), modelId)
		if err != nil {
			return fmt.Errorf("could not find model by ID \"%s\": %s", modelId, err)
		}
		modelAssertion := model.EdgeDeviceModel.ModelAssertion.AssertionBody

		// now we can build the version string
		versionString, err := device.MakeOsVersion(ctp, snaps, []byte(modelAssertion))
		if err != nil {
			return err
		}

		var result struct {
			DeviceOperatingSystemVersion string `json:"device-os-version" yaml:"device-os-version"`
		}
		result.DeviceOperatingSystemVersion = versionString
		return format.PrintFormattedOutput(cmd, result, nil)
	}
}
