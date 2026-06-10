package node

import (
	"encoding/json"
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"time"

	"github.com/spf13/cobra"
)

var changesCmd = &cobra.Command{
	Use:   "changes <device>",
	Short: "query snap changes from device",
	Args:  cobra.ExactArgs(1),
	RunE:  runChangesCmd,
}

func init() {
	snapCmd.AddCommand(changesCmd)
}

type Change struct {
	Id        string    `json:"id"`
	Kind      string    `json:"kind"`
	Summary   string    `json:"summary"`
	Status    string    `json:"status"`
	Tasks     []Task    `json:"tasks,omitempty"`
	Ready     bool      `json:"ready,omitempty"`
	SpawnTime time.Time `json:"spawn-time"`
	ReadyTime time.Time `json:"ready-time"`
	Data      string    `json:"data,omitempty"` // TODO: check type
	Err       string    `json:"err,omitempty"`
}

type Task struct {
	Id       string `json:"id"`
	Kind     string `json:"kind"`
	Summary  string `json:"summary"`
	Status   string `json:"status"`
	Progress struct {
		Label string `json:"label"`
		Done  int    `json:"done"`
		Total int    `json:"total"`
	} `json:"progress"`
	SpawnTime time.Time `json:"spawn-time"`
	ReadyTime time.Time `json:"ready-time"`
}

func runChangesCmd(cmd *cobra.Command, args []string) error {

	node := fmt.Sprintf("rpc.m2cp-gateway.%s", args[0])
	response, err := backend.DeviceRpc(cmd.Context(), node, "SnapChanges", nil)
	if err != nil {
		return err
	}

	var changes []Change
	if changesJson, ok := response.Result["changes"]; !ok {
		return fmt.Errorf("rpc response does not contain field 'changes'")
	} else {
		err = json.Unmarshal([]byte(changesJson), &changes)
		if err != nil {
			return fmt.Errorf("failed parsing changes from response: %s", err)
		}
	}

	return format.PrintFormattedOutput(cmd, changes, nil)
}
