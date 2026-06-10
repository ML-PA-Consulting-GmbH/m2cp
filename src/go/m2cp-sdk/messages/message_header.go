package messages

import (
	"crypto/sha256"
	"github.com/google/uuid"
	"m2cp"
	"m2cp/tools"
	"time"
)

type messageHeader struct {
	Type      string `json:"Type"`
	Topic     string `json:"Topic"`
	Scope     string `json:"Scope"`
	Timestamp string `json:"Timestamp"`
	Id        string `json:"Id"`
	TaskId    string `json:"TaskId"`
	Origin    string `json:"Origin"`
	via       string `json:"-"`
}

func (m *messageHeader) GetTimestamp() time.Time {
	return tools.TimeDecToTime(m.Timestamp)
}

func (m *messageHeader) GetTimestampString() string {
	return m.Timestamp
}

func (m *messageHeader) GetOrigin() string {
	return m.Origin
}

func (m *messageHeader) GetType() m2cp.MessageType {
	return newMessageType(m.Type)
}

func (m *messageHeader) GetScope() m2cp.MessageScope { return newMessageScope(m.Scope) }

func (m *messageHeader) SetScope(s m2cp.MessageScope) {
	m.Scope = s.String()
}

func (m *messageHeader) GetVia() string {
	return m.via
}

func (m *messageHeader) SetVia(via string) {
	m.via = via
}

func (m *messageHeader) GetTypeString() string {
	return m.Type
}

func (m *messageHeader) Hash() [32]byte {
	return sha256.Sum256([]byte(m.Type + "$" + m.Topic + "$" + m.Timestamp + "$" + m.Id + "$" + m.Origin))
	//fmt.Println(hex.EncodeToString(hash[:]))
}

func (m *messageHeader) SetTimestampNow() {
	m.Timestamp = tools.TimeNowDec()
}

func (m *messageHeader) SetTimestamp(t float64) {
	m.Timestamp = tools.TimeF64ToDec(t)
}

func (m *messageHeader) GetId() string {
	return m.Id
}

func (m *messageHeader) SetId(id string) {
	m.Id = id
}

func (m *messageHeader) SetIdRandom() {
	u, _ := uuid.NewRandom()
	m.Id = u.String()
}

func (m *messageHeader) GetTaskId() string {
	return m.TaskId
}

func (m *messageHeader) SetTaskId(id string) {
	m.TaskId = id
}

func (m *messageHeader) SetTaskIdRandom() {
	u, _ := uuid.NewRandom()
	m.TaskId = u.String()
}

func (m *messageHeader) GetTopic() string {
	return m.Topic
}

func (m *messageHeader) GetName() string {
	return m.Origin
}

func newMessageType(s string) m2cp.MessageType {
	switch s {
	case "COMMAND":
		return m2cp.MessageTypeCommand
	case "RESPONSE":
		return m2cp.MessageTypeResponse
	case "SIGNAL":
		return m2cp.MessageTypeSignal
	case "DATA":
		return m2cp.MessageTypeData
	case "ANY":
		return m2cp.MessageTypeAny
	case "RAW":
		return m2cp.MessageTypeRaw
	default:
		return m2cp.MessageTypeInvalid
	}
}

func newMessageScope(s string) m2cp.MessageScope {
	switch s {
	case "PROCESS":
		return m2cp.MessageScopeProcess
	case "DEVICE":
		return m2cp.MessageScopeDevice
	case "NETWORK":
		return m2cp.MessageScopeNetwork
	case "GLOBAL":
		return m2cp.MessageScopeGlobal
	default:
		return m2cp.MessageScopeUndefined
	}
}
