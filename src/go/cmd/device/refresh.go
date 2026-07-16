package device

import (
	"context"
	"fmt"
	"m2cpcli/backend"
	v5 "m2cpcli/backend/v5"
	"m2cpcli/format"
	"m2cpcli/tools"

	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh [deviceSerial or deviceName]",
	Short: "Trigger a refresh",
	Long:  `Install, change or remove apps as defined by the Deployment Group of the device`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRefreshCmd,
}

func init() {
	DeviceCmd.AddCommand(refreshCmd)
}

type RefreshResult struct {
	Messages []string `json:"messages"`
}

func (r *RefreshResult) Print(message string) {
	r.Messages = append(r.Messages, message)
}

func refreshEdgeDevice(ctx context.Context, result *RefreshResult, osSerial string) error {
	err := backend.EdgeDeviceRefresh(ctx, osSerial)
	if err != nil {
		return err
	}
	result.Print("refresh call sent to device")
	return nil
}

func refreshRealTimeDevice(ctx context.Context, result *RefreshResult, lastEdgeDeviceSerialNumber string) error {
	node := fmt.Sprintf("rpc.m2cp-coap.%s", lastEdgeDeviceSerialNumber)
	rpcCommand := "TriggerRefresh"

	result.Print(fmt.Sprintf("calling '%s' on '%s' to trigger firmware refresh of all connected Real Time Devices",
		rpcCommand, node))
	res, err := backend.DeviceRpc(ctx, node, rpcCommand, nil)
	if err != nil {
		return err
	}
	if res.Error > 0 {
		return fmt.Errorf("failed sending RPC '%s' to %s: %s (%d)", rpcCommand, node, res.Message, res.Error)
	}
	result.Print(res.Message)
	return nil
}

type DeviceModelRevisionType int

const (
	DeviceModelRevisionTypeTypeUnknown DeviceModelRevisionType = iota
	DeviceModelRevisionTypeTypeEdgeDevice
	DeviceModelRevisionTypeRealTimeDevice
)

func ParseDeviceModelRevisionType(str string) DeviceModelRevisionType {
	switch str {
	case "Edge Device":
		return DeviceModelRevisionTypeTypeEdgeDevice
	case "Real Time Device":
		return DeviceModelRevisionTypeRealTimeDevice
	case "Real-time Device":
		return DeviceModelRevisionTypeRealTimeDevice
	case "Real-Time Device":
		return DeviceModelRevisionTypeRealTimeDevice
	default:
		return DeviceModelRevisionTypeTypeUnknown
	}
}

func runRefreshCmd(cmd *cobra.Command, args []string) error {
	result := &RefreshResult{}

	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	resp, err := v5.GetDeviceInfoLimitedForRefresh(cmd.Context(), deviceId)
	if err != nil {
		return err
	}

	deviceTypeName := resp.Device.DeviceModelRevision.DeviceModel.DeviceType.DeviceTypeName
	switch ParseDeviceModelRevisionType(deviceTypeName) {
	case DeviceModelRevisionTypeTypeEdgeDevice:
		osSerial := resp.Device.SerialNumber
		err = refreshEdgeDevice(cmd.Context(), result, osSerial)
		if err != nil {
			return err
		}
	case DeviceModelRevisionTypeRealTimeDevice:
		if resp.Device.ConnectedDevice == nil {
			return fmt.Errorf("no Edge Device is known to be connected to this Real Time Device")
		}
		lastEdgeDeviceSerialNumber := resp.Device.ConnectedDevice.SerialNumber
		if tools.IsValidUuid(lastEdgeDeviceSerialNumber) {
			result.Print(fmt.Sprintf("identified most recently connected Edge Device: %s",
				lastEdgeDeviceSerialNumber))
			err = refreshRealTimeDevice(cmd.Context(), result, lastEdgeDeviceSerialNumber)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("invalid OS serial: %s", lastEdgeDeviceSerialNumber)
		}
	default:
		return fmt.Errorf("unknown device type: '%s'", deviceTypeName)
	}

	return format.PrintFormattedOutput(cmd, result, nil)
}
