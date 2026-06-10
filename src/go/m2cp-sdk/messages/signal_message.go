package messages

import (
	"fmt"
	"m2cp"
)

const (
	SignalType_Info    = "INFO"
	SignalType_Warning = "WARNING"
	SignalType_Error   = "ERROR"
	SignalType_Debug   = "DEBUG"
)

type signalMessage struct {
	Header messageHeader `json:"Header"`
	Body   signalBody    `json:"Body"`
	Trace  []trace       `json:"Trace,omitempty"`
}

type signalBody struct {
	Type    string `json:"Type"`
	Name    string `json:"Name"`
	Content string `json:"Content"`
}

func NewSignalMessage(from m2cp.Address, name, content, signalType string, scope m2cp.MessageScope) (m2cp.SignalMessage, error) {
	if signalType != SignalType_Debug && signalType != SignalType_Info && signalType != SignalType_Warning && signalType != SignalType_Error {
		return nil, fmt.Errorf("invalid level: '%s'", signalType)
	}
	if from == nil {
		return nil, fmt.Errorf("from is nil")
	}
	m := &signalMessage{
		Header: messageHeader{
			Type:   "SIGNAL",
			Topic:  "signal/" + from.GetSubtopic(),
			Scope:  scope.String(),
			Origin: from.GetAddress(),
		},
		Body: signalBody{
			Type:    signalType,
			Name:    name,
			Content: content,
		},
		Trace: []trace{},
	}
	m.Header.SetTimestampNow()
	m.Header.SetIdRandom()
	m.Header.SetTaskIdRandom()
	return m, nil
}

func (o *signalMessage) GetHeader() m2cp.MessageHeader {
	return &o.Header
}

func (o *signalMessage) GetTraces() []m2cp.MessageTrace {
	var traces = make([]m2cp.MessageTrace, len(o.Trace))
	for i := range o.Trace {
		traces[i] = &o.Trace[i]
	}
	return traces
}

func (o *signalMessage) AddTrace(relay string) {
	o.Trace = append(o.Trace, newTrace(relay))
}

func (o *signalMessage) GetType() string {
	return o.Body.Type
}

func (o *signalMessage) GetName() string {
	return o.Body.Name
}

func (o *signalMessage) GetContent() string {
	return o.Body.Content
}
