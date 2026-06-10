package node

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
)

var listCmd = &cobra.Command{
	Use:   "list [deviceSerial or deviceName]",
	Short: "list m2cp messaging nodes running on a device",
	Args:  cobra.ExactArgs(1),
	RunE:  runListCmd,
}

func init() {
	nodeCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runListCmd(cmd *cobra.Command, args []string) error {
	deviceId, err := gql.DeviceIdByNameOrSerial(cmd.Context(), args[0])
	if err != nil {
		return err
	}

	device, err := gql.DeviceByDeviceId(cmd.Context(), deviceId)
	if err != nil {
		return err
	}

	var rpcInput *gql.ExecuteRpcInput
	rpcInput = &gql.ExecuteRpcInput{
		Command: "DeviceInfo",
		Address: fmt.Sprintf("rpc.m2cp-gateway.%s", device.DeviceSerial),
	}

	var result *gql.ExecuteRpcOutput
	result, err = gql.ExecuteRpc(cmd.Context(), rpcInput)
	if err != nil {
		return fmt.Errorf("could not execute RPC: %s", err)
	}

	nodes, err := nodesFromResult(result)

	return format.PrintFormattedOutput(cmd, nodes, nil)
}

func nodesFromResult(result *gql.ExecuteRpcOutput) ([]string, error) {
	type DeviceInfoNode struct {
		Address string `json:"address"`
	}

	type DeviceInfo struct {
		IpAddresses interface{}      `json:"ipaddresses,-"`
		ListedSnaps interface{}      `json:"listedsnaps,-"`
		Users       interface{}      `json:"users,-"`
		Result      string           `json:"result"`
		Message     string           `json:"message"`
		Nodes       []DeviceInfoNode `json:"nodes"`
	}

	var deviceInfo DeviceInfo
	if len(result.Responses) > 0 {
		res := result.Responses[0].Results
		if len(res) > 0 {
			value := res[0].Value
			json.Unmarshal([]byte(value), &deviceInfo)
		}
	}

	res := make([]string, len(deviceInfo.Nodes))
	for idx, node := range deviceInfo.Nodes {
		if node.Address != "" {
			res[idx] = node.Address
		}
	}

	return res, nil
}
