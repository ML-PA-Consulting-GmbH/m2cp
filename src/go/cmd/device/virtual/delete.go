package virtual

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/docker"
	"m2cpcli/format"
	"m2cpcli/helper"
	"m2cpcli/virtual"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [containerName]",
	Short: "delete a virtual device",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runDeleteCmd,
}

func init() {
	VirtualDeviceCmd.AddCommand(deleteCmd)
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	var err error
	var cmdResultMessage string

	if !helper.UserIsLoggedIn() {
		return fmt.Errorf("user must be logged in")
	}

	err = docker.IsDockerReady()
	if err != nil {
		return err
	}

	var containerName string
	if len(args) > 0 {
		containerName = args[0]
	} else {
		containerName = virtual.GetContainerNameInteractively("Which virtual device do you want to delete?")
	}

	var deviceSerial string
	deviceSerial, err = virtual.GetDeviceSerialFromContainer(containerName)
	if err != nil {
		return err
	}

	// Now unset all fields of the device
	device, err := backend.GetDeviceWithExtendedFleetInfoBySerial(cmd.Context(), deviceSerial)
	if err != nil {
		return err
	}
	deleteDeviceFleet := shouldDeviceFleetGetDeleted(device)

	_, err = backend.SetDeviceToDead(cmd.Context(), device.Id)
	if err != nil {
		return fmt.Errorf("error unsetting device fields: %w", err)
	}

	if deleteDeviceFleet {
		_, err = backend.DeleteFleet(cmd.Context(), device.Fleet.Id)
		if err != nil {
			return fmt.Errorf("error deleting device fleet: %w", err)
		}
	}

	_, err = docker.ExecuteDockerCommand(fmt.Sprintf("rm %s --force", containerName))
	if err != nil {
		return err
	}

	type VirtualDeviceDeleteResult struct {
		Message       string
		ContainerName string
	}

	cmdResultMessage = fmt.Sprintf("Deleted virtual device '%s'", containerName)
	if deleteDeviceFleet {
		cmdResultMessage += " and its default fleet"
	}

	result := VirtualDeviceDeleteResult{
		Message:       cmdResultMessage,
		ContainerName: containerName,
	}

	return format.PrintFormattedOutput(cmd, result, nil)
}

// shouldDeviceFleetGetDeleted checks if the device fleet should be deleted depending on some conditions:
// - The fleet has to be the default fleet (name has to match device name, description has to be the default description)
// - The fleet has to contain only the device
func shouldDeviceFleetGetDeleted(device *backend.GetDevicesWithExtendedFleetInfoBySerialResultEdgeDeviceCollectionSegmentItemsEdgeDevice) bool {
	if device.Fleet == nil {
		return false
	}

	// Name of the fleet has to be the same as the device name, otherwise the fleet was probably created manually
	if device.Fleet.FleetName != *device.DeviceName {
		return false
	}

	// Description of the fleet has to be the default description
	if *device.Fleet.Description != VirtualDeviceDefaultFleetDescription {
		return false
	}

	// The device has to be the only device in the fleet
	if len(device.Fleet.EdgeDevices) != 1 && device.Fleet.EdgeDevices[0].Id != device.Id {
		return false
	}

	return true
}
