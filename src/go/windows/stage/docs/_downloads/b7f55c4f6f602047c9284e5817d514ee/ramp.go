package main

import (
	"fmt"
	"m2cp"
	"m2cp/m2cp_new"
	"m2cp/messages"
	"m2cp/networks"
	"m2cp/rpc"
	"m2cp/rpc/rpctypes"
	"os"
	"time"
)

var period = 5

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func getPeriod() m2cp.RpcCommand {
	cmd, err := rpc.NewRpcCommand(
		"GetPeriod",
		"Get the current ramp period.",
		"Returns a single integer.",
		nil,
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeInt, "period", "the current period length", false),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			res := rpctypes.NewRpcResultSuccess("Success")
			res.SetInt("period", period)
			return res
		})
	handleError(err)
	return cmd
}

func setPeriod() m2cp.RpcCommand {
	cmd, err := rpc.NewRpcCommand(
		"SetPeriod",
		"Set the ramp period.",
		"",
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeInt, "period", "the new period length", false),
		},
		nil,
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			period = parameters.GetInt("period", 5)
			res := rpctypes.NewRpcResultSuccess("Success")
			return res
		})
	handleError(err)
	return cmd
}

func main() {
	context := m2cp_new.ContextPlus()
	context = context.SetModule("ramp-main")
	context.SetLogLevel(m2cp.LogLevelInfo)
	context.LogInfo("Ramp! Using SDK version: %s (Strg+c to exit)", m2cp.GetVersion())

	connection, err := networks.NewNetworkConnection(context)
	handleError(err)
	defer connection.Close()

	node, err := connection.NewNode("ramp")
	handleError(err)
	node.SetDataQueueMaxCount(0)
	context.LogInfo("Node address: %s", node.GetAddress())

	handler, err := rpc.NewRpcHandler(
		getPeriod(),
		setPeriod())
	handleError(err)
	node.SetRpcHandler(handler)

	const name = "number"
	fields := []m2cp.DataField{
		messages.NewDataFieldInt(name),
	}
	format, err := messages.NewDataFormat("a673ab02-160a-4db9-a44e-71e82b3c4235", fields...)
	handleError(err)

	for {
		for counter := 0; counter < period; counter++ {
			row, err := format.NewRow(map[string]interface{}{
				name: counter,
			})
			handleError(err)
			err = node.EmitDataRow(row)
			handleError(err)

			time.Sleep(1 * time.Second)
			fmt.Print(".")
		}
	}
}
