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
	"strconv"
	"time"
)

const appName = "demo-square"

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func squareOf() m2cp.RpcCommand {
	cmd, err := rpc.NewRpcCommand(
		"SquareOf",
		"Calculate the square of a number",
		"Base and result of the square operation",
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeInt, "base", "The number to be squared", false),
		},
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeInt, "square", "The squared number", false),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			number := parameters.GetInt("base", 0)
			square := number * number
			res := rpctypes.NewRpcResultSuccess("Success")
			res.SetInt("square", square)
			return res
		},
	)
	handleError(err)
	return cmd
}

func main() {
	context := m2cp_new.ContextPlus()
	context = context.SetModule("square-main")
	context.SetLogLevel(m2cp.LogLevelInfo)
	context.LogInfo("Square! Using SDK version: %s (Strg+c to exit)", m2cp.GetVersion())

	connection, err := networks.NewNetworkConnectionWithOptions(context, m2cp.NetworkConnectionOptions{
		AppName: appName,
	})
	handleError(err)
	defer connection.Close()

	node, err := connection.NewNode("rpc")
	handleError(err)

	handleData := func(data m2cp.DataMessage) {
		for _, row := range data.GetRows() {
			number := *row.GetFieldInt("number")
			numberSquared := number * number
			context.LogInfo(fmt.Sprintf("square(%v) = %v", number, numberSquared))

			err = node.EmitSignal("number", strconv.Itoa(numberSquared), messages.SignalType_Info)
			handleError(err)
		}
	}
	const topic = "ramp/"
	err = connection.SubscribeData([]string{topic}, handleData)
	handleError(err)

	handler, err := rpc.NewRpcHandler(
		squareOf(),
	)
	handleError(err)
	node.SetRpcHandler(handler)

	for {
		select {
		case <-context.Done():
			return
		default:
			time.Sleep(1 * time.Second)
			fmt.Print(".")
		}

	}
}
