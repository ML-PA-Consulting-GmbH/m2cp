package node

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	"m2cpcli/graphql"
	"m2cpcli/helper"
)

var statemachineCmd = &cobra.Command{
	Use:   "statemachine",
	Short: "control a statemachine running at a given address",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatemachineCmd,
}

func init() {
	nodeCmd.AddCommand(statemachineCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	//statemachineCmd.Flags().StringP("address", "a", "", "m2cp messaging address of statemachine node")
	statemachineCmd.Flags().StringP("command", "c", "", "command name (state, states, logs, trigger, jump, resetCounter, counters)")
	err := statemachineCmd.MarkFlagRequired("command")
	cobra.CheckErr(err)
	statemachineCmd.Flags().StringP("target", "t", "", "target state name for commands 'trigger', 'jump' and 'resetCounter'")
}

func runStatemachineCmd(cmd *cobra.Command, args []string) error {
	var err error
	var rpcInput *graphql.ExecuteRpcInput

	// TODO: error when not logged in!

	rpcInput, err = newStateMachineRpcInput(cmd, args)
	if err != nil {
		return fmt.Errorf("invalid statemachine input: %s", err)
	}

	err = helper.ValidateRpcInput(rpcInput)
	if err != nil {
		return fmt.Errorf("invalid input: %s", err)
	}

	// cmd.Println(rpcInput) // TODO: remove this debug printing

	result, err := graphql.ExecuteRpc(cmd.Context(), rpcInput)
	if err != nil {
		return fmt.Errorf("could not execute RPC: %s", err)
	}

	// TODO: add message for user?
	//var RpcResult struct {
	//	Message string      `json:"message"`
	//	Foo     interface{} `json:"foo"`
	//}
	//RpcResult.Message = "ok" // TODO: if all errors are 0
	//RpcResult.Foo = result

	return format.PrintFormattedOutput(cmd, result, nil)
}

// translateCliCommandsToRpcCommands trys to translate else returns input
func translateCliCommandToRpcCommand(command string) string {
	result := ""
	switch command {
	case "state":
		result = "GetCurrentState"
	case "trigger":
		result = "TriggerExecution"
	case "jump":
		result = "JumpToState"
	case "states":
		result = "GetStates"
	case "logs":
		result = "GetLogs"
	case "counters":
		result = "GetActionExecutionCounters"
	case "resetCounter":
		result = "ResetActionExecutionCounter"
	default:
		return command
	}
	return result
}

// newStateMachineRpcInput collects the command name and additional parameters as needed.
func newStateMachineRpcInput(cmd *cobra.Command, args []string) (*graphql.ExecuteRpcInput, error) {
	rpcInput := graphql.ExecuteRpcInput{}
	rpcInput.Address = args[0]

	const commandArg = "command"
	if cmd.Flags().Changed(commandArg) {
		cmdName, err := cmd.Flags().GetString(commandArg)
		if err != nil {
			return nil, fmt.Errorf("could not get %s argument", commandArg)
		}
		rpcInput.Command = translateCliCommandToRpcCommand(cmdName) // the statemachine checks command names for itself

		if rpcInput.Command == "TriggerExecution" || rpcInput.Command == "JumpToState" || rpcInput.Command == "ResetActionExecutionCounter" {
			var stateName string
			stateName, err = cmd.Flags().GetString("target")
			if err != nil {
				return nil, fmt.Errorf("could not get state argument")
			}
			if stateName == "" {
				return nil, fmt.Errorf("state must not be empty")
			}
			rpcInput.Parameters = []graphql.ExecuteRpcParameterInput{
				{
					Key:   "state",
					Value: stateName,
				},
			}
		} else {
			// the other commands require no arguments
		}
	}
	return &rpcInput, nil
}
