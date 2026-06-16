package main

import (
	"fmt"
	"m2cp"
	"m2cp/m2cp_new"
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
	context = context.SetModule("subscribe-main")
	context.SetLogLevel(m2cp.LogLevelInfo)
	context.LogInfo("Subscribe to any SIGNAL! Using SDK version: %s (Strg+c to exit)", m2cp.GetVersion())

	connection, err := networks.NewNetworkConnection(context)
	handleError(err)
	defer connection.Close()

	handleSignal := func(signal m2cp.SignalMessage) {
		name := signal.GetName()
		value := signal.GetContent()
		level := signal.GetType()
		context.LogInfo(fmt.Sprintf("%s: %s, %s.", level, name, value))
	}
	err = connection.SubscribeSignals([]string{"greeter.#"}, handleSignal)
	handleError(err)

	for {
		select {
		case <-context.Done():
			return
		default:
			context.Sleep(1 * time.Second)
			fmt.Print(".")
		}
	}
}
