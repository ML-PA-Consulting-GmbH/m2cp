package messages

import (
	"m2cp"
)

type commandMessage struct {
	Header messageHeader `json:"Header"`
	Body   commandBody   `json:"Body"`
	Trace  []trace       `json:"Trace,omitempty"`
}

type commandBody struct {
	Command    string
	Parameters []map[string]string
}

func NewCommandMessage(from, to m2cp.Address, command string) (m2cp.CommandMessage, error) {
	m := &commandMessage{
		Header: messageHeader{
			Type:   "COMMAND",
			Topic:  "command/" + to.GetAddress(),
			Scope:  "GLOBAL",
			Origin: from.GetAddress(),
		},
		Body: commandBody{
			Command: command,
		},
	}
	m.Header.SetTimestampNow()
	m.Header.SetIdRandom()
	m.Header.SetTaskIdRandom()
	return m, nil
}

func (o *commandMessage) SetParameters(parameters []map[string]string) {
	o.Body.Parameters = parameters
}

func (o *commandMessage) GetParameters() []map[string]string {
	return o.Body.Parameters
}

func (o *commandMessage) GetCommand() string {
	return o.Body.Command
}

func (o *commandMessage) GetHeader() m2cp.MessageHeader {
	return &o.Header
}

func (o *commandMessage) GetTraces() []m2cp.MessageTrace {
	var traces = make([]m2cp.MessageTrace, len(o.Trace))
	for i := range o.Trace {
		traces[i] = &o.Trace[i]
	}
	return traces
}

func (o *commandMessage) AddTrace(relay string) {
	o.Trace = append(o.Trace, newTrace(relay))
}

func (o *commandMessage) SetTraces(traces []m2cp.MessageTrace) {
	o.Trace = make([]trace, len(traces))
	for i, t := range traces {
		o.Trace[i] = trace{
			A: t.GetRelay(),
			T: t.GetTimeString(),
		}
	}
}

func (o *commandMessage) AddParameters(parameters map[string]string) {
	if o.Body.Parameters == nil {
		o.Body.Parameters = []map[string]string{}
	}
	o.Body.Parameters = append(o.Body.Parameters, parameters)
}
