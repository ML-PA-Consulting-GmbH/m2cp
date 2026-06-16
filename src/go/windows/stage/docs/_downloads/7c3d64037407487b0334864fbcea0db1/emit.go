package main

import (
	"fmt"
	"m2cp"
	"m2cp/m2cp_new"
	"m2cp/messages"
	"m2cp/networks"
	"os"
	"time"
)

func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	context := m2cp_new.ContextPlus()
	context = context.SetModule("emit-main")
	context.SetLogLevel(m2cp.LogLevelInfo)
	context.LogInfo("Hello, this is the m2cp-tutorial. Using SDK version: %s", m2cp.GetVersion())

	connection, err := networks.NewNetworkConnection(context)
	handleError(err)
	defer connection.Close()

	node, err := connection.NewNode("greeter")
	handleError(err)

	name := "message"
	value := "Hello, World!"
	level := messages.SignalType_Info
	err = node.EmitSignal(name, value, level)
	handleError(err)

	context.Sleep(1 * time.Second)
}
