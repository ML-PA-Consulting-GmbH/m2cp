package node

import (
	"encoding/json"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"

	"github.com/spf13/cobra"
)

var changeCmd = &cobra.Command{
	Use:   "change <device> <change-id>",
	Short: "query details on a snap change from device",
	Args:  cobra.ExactArgs(2),
	RunE:  runChangeCmd,
}

func init() {
	snapCmd.AddCommand(changeCmd)
}

func runChangeCmd(cmd *cobra.Command, args []string) error {

	node := fmt.Sprintf("rpc.m2cp-gateway.%s", args[0])
	response, err := backend.DeviceRpc(cmd.Context(), node, "SnapChange", map[string]string{
		"id": args[1],
	})
	if err != nil {
		return err
	}

	var change Change
	if changesJson, ok := response.Result["change"]; !ok {
		return fmt.Errorf("rpc response does not contain field 'change'")
	} else {
		err = json.Unmarshal([]byte(changesJson), &change)
		if err != nil {
			return fmt.Errorf("failed parsing changes from response: %s", err)
		}
	}

	return format.PrintFormattedOutput(cmd, change, nil)
}
