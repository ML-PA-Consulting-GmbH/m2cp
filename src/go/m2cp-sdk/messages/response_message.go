package messages

import (
	"encoding/base64"
	"fmt"
	"m2cp"
	"m2cp/rpc/rpctypes"
)

type responseMessage struct {
	Header messageHeader `json:"Header"`
	Body   responseBody  `json:"Body"`
	Trace  []trace       `json:"Trace,omitempty"`
}

type responseBody struct {
	CommandId string
	Responses []response
}

type response struct {
	Error   int
	Message string
	Results map[string]string
}

func NewResponseMessage(cm m2cp.CommandMessage) (m2cp.ResponseMessage, error) {
	if cm == nil {
		return nil, fmt.Errorf("parameter is nil")
	}
	if len(cm.GetHeader().GetTopic()) < 9 {
		return nil, fmt.Errorf("invalid origin: '%s'", cm.GetHeader().GetTopic())
	}

	cmTraces := cm.GetTraces()

	m := &responseMessage{
		Header: messageHeader{
			Type:   "RESPONSE",
			Topic:  "response/" + cm.GetHeader().GetName(),
			Scope:  cm.GetHeader().GetScope().String(),
			Origin: cm.GetHeader().GetTopic()[8:],
		},
		Body: responseBody{
			CommandId: cm.GetHeader().GetId(),
			Responses: []response{},
		},
		Trace: make([]trace, len(cmTraces)),
	}

	for i := range cmTraces {
		cmTrace := cmTraces[i]
		m.Trace[i] = trace{A: cmTrace.GetRelay(), T: cmTrace.GetTimeString()}
	}

	m.GetHeader().SetTimestampNow()
	m.GetHeader().SetIdRandom()
	m.GetHeader().SetTaskId(cm.GetHeader().GetTaskId())
	m.GetHeader().SetVia(cm.GetHeader().GetVia())
	return m, nil
}

func NewResponseMessageFailed(cm m2cp.CommandMessage, errorMessage string) (m2cp.ResponseMessage, error) {
	m, err := NewResponseMessage(cm)
	if err != nil {
		return nil, err
	}
	m.AddResult(rpctypes.NewRpcResultSuccess(errorMessage))
	return m, nil
}

func (o *responseMessage) GetHeader() m2cp.MessageHeader {
	return &o.Header
}

func (o *responseMessage) GetTraces() []m2cp.MessageTrace {
	var traces = make([]m2cp.MessageTrace, len(o.Trace))
	for i := range o.Trace {
		traces[i] = &o.Trace[i]
	}
	return traces
}

func (o *responseMessage) AddTrace(relay string) {
	o.Trace = append(o.Trace, newTrace(relay))
}

func (o *responseMessage) SetTraces(traces []m2cp.MessageTrace) {
	o.Trace = make([]trace, len(traces))
	for i, t := range traces {
		o.Trace[i] = trace{
			A: t.GetRelay(),
			T: t.GetTimeString(),
		}
	}
}

func (o *responseMessage) Type() string {
	return o.Header.Type
}

func (o *responseMessage) GetCommandId() string {
	return o.Body.CommandId
}

func (o *responseMessage) CountResults() int {
	return len(o.Body.Responses)
}

func (o *responseMessage) HasResponses() bool {
	return o.Body.Responses != nil && len(o.Body.Responses) > 0
}

func (o *responseMessage) GetResult(index int) (m2cp.RpcResultReadonly, error) {
	if index < 0 || index >= len(o.Body.Responses) {
		return nil, fmt.Errorf("index out of range")
	}
	res := o.Body.Responses[index]
	return &rpctypes.RpcObject{
		Dict:    res.Results,
		Error:   m2cp.RpcErrorCode(res.Error),
		Message: res.Message,
		Trace:   o.GetTraces(),
	}, nil
}

func (o *responseMessage) GetResults() []m2cp.RpcResultReadonly {
	rs := make([]m2cp.RpcResultReadonly, len(o.Body.Responses))
	for i, res := range o.Body.Responses {
		rs[i] = &rpctypes.RpcObject{
			Dict:    res.Results,
			Error:   m2cp.RpcErrorCode(res.Error),
			Message: res.Message,
			Trace:   o.GetTraces(),
		}
	}
	return rs
}

func (o *responseMessage) SetResults(results []m2cp.RpcResult) {
	rs := make([]response, len(results))
	for i, r := range results {
		rs[i] = response{
			Error:   int(r.GetErrorCode()),
			Message: r.GetMessage(),
			Results: r.GetAllRaw(),
		}
	}
	o.Body.Responses = rs
}

func (o *responseMessage) AddResult(result m2cp.RpcResult) {
	r := response{
		Error:   int(result.GetErrorCode()),
		Message: result.GetMessage(),
		Results: result.GetAllRaw(),
	}
	o.Body.Responses = append(o.Body.Responses, r)
}

func (o *responseMessage) HasErrors() bool {
	if o.Body.Responses == nil {
		return false
	}
	for _, res := range o.Body.Responses {
		if res.Error != 0 {
			return true
		}
	}
	return false
}

func (o *responseMessage) GetErrors() []string {
	if o.Body.Responses == nil {
		return []string{}
	}
	errors := []string{}
	for _, res := range o.Body.Responses {
		if res.Error != 0 {
			errors = append(errors, res.Message)
		}
	}
	return errors
}

func (r *response) GetJson(field string) (string, error) {
	val, ok := r.Results[field]
	if !ok {
		return "", fmt.Errorf("field %s not found in response", field)
	}
	decoded, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func (r *response) GetError() int {
	return r.Error
}

func (r *response) GetMessage() string {
	return r.Message
}

func (r *response) GetResults() map[string]string {
	return r.Results
}
