package node

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"m2cpcli/format"
	"m2cpcli/graphql"
	gql "m2cpcli/graphql"
	"m2cpcli/helper"
	"m2cpcli/tools"
	"os"
	"strings"
)

type RpcResult struct {
	DeviceSerial string                             `json:"device-serial" yaml:"device-serial,omitempty"`
	Command      string                             `json:"command" yaml:"command,omitempty"`
	CommandId    string                             `json:"command-id" yaml:"command-id,omitempty"`
	Parameters   []graphql.ExecuteRpcParameterInput `json:"parameters" yaml:"parameters,omitempty"`
	Responses    []RpcResultResponse                `json:"responses" yaml:"response"`
}

type RpcResultResponse struct {
	Error   int               `json:"error" yaml:"error"`
	Message string            `json:"message" yaml:"message"`
	Results map[string]string `json:"results" yaml:"results"`
}

var rpcCmd = &cobra.Command{
	Use:   "rpc [nodeAddress]",
	Short: "send an RPC command to node",
	Long:  "Send an RPC command to a node. The command and parameters can be provided as a JSON object via stdin and as command line parameters. Command line arguments and flags have precedence over stdin.",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runRpcCmd,
}

func init() {
	nodeCmd.AddCommand(rpcCmd)

	// rpcCmd.Flags().StringP("address", "a", "", "m2cp messaging address of target node")
	rpcCmd.Flags().StringP("command", "c", "", "RPC command name")
	rpcCmd.Flags().Bool("stdin", false, "read JSON of RPC command body from stdin")
	rpcCmd.Flags().StringArrayP("param", "p", []string{}, "key-value pair for parameters (can be specified multiple times) specified as key=value")
	rpcCmd.Flags().BoolP("verbose", "v", false, "print verbose output including resolved OS Serial Number and the used command and parameters")

	// verify mutual exclusion and that either stdin or command is provided
	rpcCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		flagFromStdin := cmd.Flags().Changed("stdin")
		flagCommand := cmd.Flags().Changed("command")

		if !flagFromStdin && !flagCommand {
			return errors.New("either stdin or command flag must be provided")
		}

		return nil
	}

	//err := rpcCmd.MarkFlagRequired("command")
	// cobra.CheckErr(err)
	//rpcCmd.Flags().String("device-hub", "", "development: use alternative device hub")

}

func runRpcCmd(cmd *cobra.Command, args []string) error {
	var readRpcFromStdin bool
	var err, warning error
	var rpcInput *graphql.ExecuteRpcInput
	rpcInput = &graphql.ExecuteRpcInput{}

	var resultOutput *RpcResult
	resultOutput = &RpcResult{}

	// This way, the command can be either defined in the JSON or as a command line parameter.
	// Command line has precedence.
	readRpcFromStdin = cmd.Flags().Changed("stdin")
	if readRpcFromStdin {
		warning = overwriteRpcInputFromStdin(rpcInput)
		if warning != nil {
			cmd.Println(fmt.Errorf("warning: %s", warning))
		}
	}
	warning = overwriteRpcInputFromCliArgs(rpcInput, cmd, args)
	if warning != nil {
		cmd.Println(fmt.Errorf("warning: %s", warning))
	}

	err = resolveDeviceNameToSerial(rpcInput, cmd, resultOutput)
	if err != nil {
		return fmt.Errorf("invalid address: %s", err)
	}

	err = helper.ValidateRpcInput(rpcInput)
	if err != nil {
		return fmt.Errorf("invalid input: %s", err)
	}

	var rpcResult *graphql.ExecuteRpcOutput
	rpcResult, err = graphql.ExecuteRpc(cmd.Context(), rpcInput)
	if err != nil {
		return fmt.Errorf("could not execute RPC: %s", err)
	}

	resultOutput.CommandId = rpcResult.CommandId
	resultOutput.Command = rpcInput.Command
	resultOutput.Responses = make([]RpcResultResponse, len(rpcResult.Responses))
	for i, r := range rpcResult.Responses {
		res := RpcResultResponse{
			Error:   r.Error,
			Message: r.Message,

			Results: map[string]string{},
		}
		for _, v := range r.Results {
			res.Results[v.Key] = v.Value
		}
		resultOutput.Responses[i] = res
	}

	if cmd.Flags().Changed("verbose") {
		resultOutput.Command = rpcInput.Command
		resultOutput.Parameters = rpcInput.Parameters
	}

	return format.PrintFormattedOutput(cmd, resultOutput, nil)
}

// overwriteRpcInputFromStdin changes the data, where `rpcInput` points to
func overwriteRpcInputFromStdin(rpcInput *graphql.ExecuteRpcInput) error {
	input, err := readStdinPipe()
	if err != nil {
		return fmt.Errorf("could not read from stdin: %s", err)
	}

	// The JSON that a user provides differs from what we need in GraphQL, so we transform it.
	type UserRpcInput struct {
		Command    string              `json:"Command"`
		Parameters []map[string]string `json:"Parameters"`
	}
	var userRpcInput *UserRpcInput

	err = json.Unmarshal(input, &userRpcInput)
	if err != nil {
		return fmt.Errorf("invalid user json input: %s", err)
	}

	rpcInput.Command = userRpcInput.Command
	for _, dict := range userRpcInput.Parameters {
		for key, value := range dict {
			newParam := graphql.ExecuteRpcParameterInput{
				Key:   key,
				Value: value,
			}
			rpcInput.Parameters = append(rpcInput.Parameters, newParam)
		}
	}
	return nil
}

// overwriteRpcInputFromCliArgs changes the data, where `rpcInput` points to
func overwriteRpcInputFromCliArgs(rpcInput *graphql.ExecuteRpcInput, cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		rpcInput.Address = args[0]
	}

	const commandArg = "command"
	if cmd.Flags().Changed(commandArg) {
		cmdName, err := cmd.Flags().GetString(commandArg)
		if err != nil {
			return fmt.Errorf("could not get command argument")
		}
		rpcInput.Command = cmdName
	}
	const paramArg = "param"
	params, err := cmd.Flags().GetStringArray(paramArg)
	// number of parameters can be zero
	if err != nil {
		return nil
	}

	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("param must be in the format key=value")
		}

		newParam := graphql.ExecuteRpcParameterInput{
			Key:   kv[0],
			Value: kv[1],
		}

		// if the key already exists, overwrite it
		index, _, _ := helper.FindFirst(
			rpcInput.Parameters,
			func(p graphql.ExecuteRpcParameterInput) bool {
				return p.Key == newParam.Key
			})

		if rpcInput.Parameters == nil {
			rpcInput.Parameters = []graphql.ExecuteRpcParameterInput{}
		}
		if index >= 0 {
			rpcInput.Parameters[index] = newParam
		} else {
			rpcInput.Parameters = append(rpcInput.Parameters, newParam)
		}
	}
	return nil
}

func resolveDeviceNameToSerial(rpcInput *graphql.ExecuteRpcInput, cmd *cobra.Command, result *RpcResult) error {

	if rpcInput.Address == "" {
		return fmt.Errorf("address missing")
	}

	address := strings.Split(rpcInput.Address, ".")
	if len(address) != 3 {
		return fmt.Errorf("address should be in fromat '<node>.<app-name>.<device-serial>' or '<node>.<app-name>.<device-name>'")
	}
	if tools.IsValidUuid(address[2]) {
		result.DeviceSerial = address[2]
	} else {
		deviceSerial, err := gql.DeviceSerialByName(cmd.Context(), address[2])
		if err != nil {
			return err
		}
		result.DeviceSerial = deviceSerial
		rpcInput.Address = fmt.Sprintf("%s.%s.%s", address[0], address[1], deviceSerial)
	}
	return nil
}

func readStdinPipe() ([]byte, error) {
	nBytes, nChunks := int64(0), int64(0)
	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 0, 4*1024)
	var input []byte
	for {
		n, err := r.Read(buf[:cap(buf)])
		buf = buf[:n]
		if n == 0 {
			if err == nil {
				continue
			}
			if err == io.EOF {
				break
			}
			return nil, err
		}
		nChunks++
		nBytes += int64(len(buf))
		input = append(input, buf...)
		if err != nil && err != io.EOF {
			return nil, err
		}
	}

	return input, nil
}
