package rpc

import (
	"fmt"
	"m2cp"
	"m2cp/messages"
	"m2cp/rpc/rpctypes"
	"strings"
)

type rpcHandler struct {
	systemCommands map[string]m2cp.RpcCommand
	commands       map[string]m2cp.RpcCommand
}

func NewRpcHandler(commands ...m2cp.RpcCommand) (m2cp.RpcHandler, error) {
	commandsByName := make(map[string]m2cp.RpcCommand, len(commands))
	for _, c := range commands {
		if c == nil {
			continue
		}
		if _, ok := commandsByName[c.GetName()]; ok {
			return nil, fmt.Errorf("command name already exists: %s", c.GetName())
		}
		commandsByName[c.GetName()] = c
	}
	ch := &rpcHandler{
		commands: commandsByName,
	}
	ch.systemCommands = map[string]m2cp.RpcCommand{
		"NodeMemoryUsage":   cmdNodeMemoryUsage(),
		"NodeUptime":        cmdNodeUptime(),
		"NodeDocumentation": cmdNodeDocumentation(ch),
	}
	return ch, nil
}

func (r *rpcHandler) Process(ctpParent m2cp.ContextPlus, commandMessage m2cp.CommandMessage) (m2cp.ResponseMessage, error) {
	resMsg, err := messages.NewResponseMessage(commandMessage)
	if err != nil {
		return nil, err
	}
	resMsg.AddTrace("execute")
	cmdName := commandMessage.GetCommand()
	ctp := ctpParent.BranchWithName("rpc:" + cmdName)
	cmd, ok := r.systemCommands[cmdName]
	if !ok {
		cmd, ok = r.commands[cmdName]
		if !ok {
			responses := make([]m2cp.RpcResult, len(commandMessage.GetParameters()))
			for i := range responses {
				responses[i] = rpctypes.NewRpcResultError("command not found", m2cp.RpcErrorCodeCommandNotFound)
			}
			resMsg.SetResults(responses)
			return resMsg, nil
		}
	}
	paramSets := commandMessage.GetParameters()
	if len(paramSets) == 0 {
		res := cmd.Execute(ctp, nil)
		resMsg.AddResult(res)
	} else {
		for _, ps := range paramSets {
			res := cmd.Execute(ctp, ps)
			resMsg.AddResult(res)
		}
	}
	return resMsg, nil
}

func (r *rpcHandler) GetDocumentation() []m2cp.RpcDocumentation {
	docs := make([]m2cp.RpcDocumentation, len(r.systemCommands)+len(r.commands))
	i := 0
	for _, c := range r.systemCommands {
		docs[i] = c.GetDocumentation()
		i++
	}
	for _, c := range r.commands {
		docs[i] = c.GetDocumentation()
		i++
	}
	return docs
}

func (r *rpcHandler) String() string {
	docs := r.GetDocumentation()
	var sb strings.Builder
	for _, doc := range docs {
		sb.WriteString("\n")
		sb.WriteString(string(doc.ToJson()))
	}
	return sb.String()
}
