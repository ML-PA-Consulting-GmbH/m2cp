package rpc

import (
	"fmt"
	"m2cp"
	"m2cp/rpc/rpctypes"
	"runtime"
	"strings"
	"time"
)

func cmdNodeUptime() m2cp.RpcCommand {
	startTime := time.Now()
	cmd, err := NewRpcCommand(
		"NodeUptime",
		"Get the current uptime of this plugin",
		"Get uptime of the node",
		[]m2cp.RpcParameter{},
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "uptime", "current uptime of this plugin in seconds", false),
			NewRpcParameter(m2cp.TypeString, "sdk", "sdk version used to build this app", false),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			res := rpctypes.NewRpcResultSuccess("ok")
			res.SetLong("uptime", int64(time.Since(startTime).Seconds()))
			res.SetString("sdk", m2cp.GetVersion())
			return res
		},
	)
	if err != nil {
		panic(fmt.Sprintf("system command 'NodeUptime' is invalid: %s", err.Error()))
	}
	return cmd
}

func cmdNodeMemoryUsage() m2cp.RpcCommand {
	cmd, err := NewRpcCommand(
		"NodeMemoryUsage",
		"Get the current memory in MBytes used by this plugin",
		"Get memory usage of the node",
		[]m2cp.RpcParameter{},
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "memoryUsageAct", "current memory usage in MBytes used by this plugin", false),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			res := rpctypes.NewRpcResultSuccess("ok")
			res.SetLong("memoryUsageAct", int64(getMemoryUsageInMb()))
			return res
		},
	)
	if err != nil {
		panic(fmt.Sprintf("system command 'NodeMemoryUsage' is invalid: %s", err.Error()))
	}
	return cmd
}

func cmdNodeDocumentation(rpcHandler m2cp.RpcHandler) m2cp.RpcCommand {
	cmd, err := NewRpcCommand(
		"NodeDocumentation",
		"Get a list of all available RPC commands of this node together with their call and return schemes.",
		"Get documentation of all RPC commands",
		[]m2cp.RpcParameter{},
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeString, "documentation", "json object with all available RPC commands of this node", false),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			docs := rpcHandler.GetDocumentation()
			docsJoined := make([]string, len(docs))
			for i, doc := range docs {
				docsJoined[i] = string(doc.ToJson())
			}
			docsString := "[" + strings.Join(docsJoined, ",") + "]"
			res := rpctypes.NewRpcResultSuccess("ok")
			res.SetBinary("documentation", []byte(docsString))
			return res
		},
	)
	if err != nil {
		panic(fmt.Sprintf("system command 'NodeDocumentation' is invalid: %s", err.Error()))
	}
	return cmd
}

func getMemoryUsageInMb() uint64 {
	memUsage := new(runtime.MemStats)
	runtime.ReadMemStats(memUsage)
	return memUsage.Alloc / 1024 / 1024
}
