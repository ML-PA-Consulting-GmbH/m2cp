package device

import (
	"context"
	"encoding/json"
	"fmt"
	coap_server "m2cp/coap/coap-server"
	"m2cpcli/backend"
	"m2cpcli/format"
	"strconv"

	"github.com/spf13/cobra"
)

var peersCmd = &cobra.Command{
	Use:   "peers [deviceSerial or deviceName]",
	Short: "Report known peers",
	Long:  `Send a RPC to an Edge Device to report its known peers state (this is an unfiltered internal state, useful for debugging purposes but potentially hard to read).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPeersCmd,
}

func init() {
	DeviceCmd.AddCommand(peersCmd)
	peersCmd.Flags().BoolP("unfiltered", "u", false, "unfiltered peers list (default: false)")
}

// PeersType is exported so the coap sub-package can reuse it.
type PeersType []struct {
	Peer coap_server.Peer `json:"peer"`
}

func runPeersCmd(cmd *cobra.Command, args []string) error {
	deviceId, err := backend.GetDeviceIdByIdOrNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	device, err := backend.GetDeviceInfoById(cmd.Context(), deviceId)
	if err != nil {
		return err
	}

	if !device.IsED() {
		return fmt.Errorf("this command only supports Edge Devices")
	}
	unfiltered, _ := cmd.Flags().GetBool("unfiltered")

	peersMap, err := FetchPeersMap(cmd.Context(), device.DeviceSerial, unfiltered)
	if err != nil {
		return err
	}

	var peersList []coap_server.Peer
	for _, peer := range peersMap {
		peersList = append(peersList, peer.Peer)
	}
	return format.PrintFormattedOutput(cmd, peersList, nil)
}

// FetchPeersMap is exported so the coap sub-package can reuse it.
func FetchPeersMap(ctx context.Context, deviceSerial string, unfiltered bool) (PeersType, error) {
	node := fmt.Sprintf("rpc.m2cp-coap.%s", deviceSerial)
	rpcCommand := "listKnownPeers"
	res, err := backend.DeviceRpc(ctx, node, rpcCommand, map[string]string{"unfiltered": strconv.FormatBool(unfiltered)})
	if err != nil {
		return nil, err
	}

	var peersMap PeersType

	if list, ok := res.Result["list"]; !ok {
		return nil, fmt.Errorf("list not found in response: %v", res)
	} else if err = json.Unmarshal([]byte(list), &peersMap); err != nil {
		return nil, fmt.Errorf("failed parsing received list (%s). raw rpc result:\n %s", err, list)
	}

	return peersMap, nil
}
