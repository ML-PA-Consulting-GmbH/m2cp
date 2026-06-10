package node

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
)

var infoCmd = &cobra.Command{
	Use:   "info [nodeAddress]",
	Short: "get information about a node in the m2cp messaging network",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfoCmd,
}

func init() {
	nodeCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runInfoCmd(cmd *cobra.Command, args []string) error {

	var rpcInput *gql.ExecuteRpcInput
	rpcInput = &gql.ExecuteRpcInput{
		Command: "NodeDocumentation",
		Address: args[0],
	}

	var result *gql.ExecuteRpcOutput
	result, err := gql.ExecuteRpc(cmd.Context(), rpcInput)
	if err != nil {
		return fmt.Errorf("could not execute RPC: %s", err)
	}

	var jsonString []byte
	if len(result.Responses) > 0 {
		res := result.Responses[0].Results
		if len(res) > 0 {
			value := res[0].Value
			jsonString, err = base64.StdEncoding.DecodeString(value)
			if err != nil {
				return err
			}
		}
	}

	type NodeDocumentation struct {
		Name               string      `json:"name"`
		Type               string      `json:"type"`
		Description        string      `json:"description"`
		DescriptionReturns string      `json:"description-returns" yaml:"description-returns"`
		Parameters         interface{} `json:"parameters"`
		Returns            interface{} `json:"returns"`
	}

	var doc []NodeDocumentation
	err = json.Unmarshal(jsonString, &doc)
	if err != nil {
		return err
	}

	return format.PrintFormattedOutput(cmd, doc, nil)
}
