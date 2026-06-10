package m2cp

import (
	"m2cp/messages/pb"
	"time"
)

type MessageType int32

const (
	MessageTypeCommand  = MessageType(pb.MessageType_COMMAND)
	MessageTypeResponse = MessageType(pb.MessageType_RESPONSE)
	MessageTypeSignal   = MessageType(pb.MessageType_SIGNAL)
	MessageTypeData     = MessageType(pb.MessageType_DATA)
	MessageTypeInvalid  = MessageType(pb.MessageType_INVALID)
	MessageTypeAny      = MessageType(pb.MessageType_ANY)
	MessageTypeRaw      = MessageType(pb.MessageType_RAW)
)

func (m MessageType) String() string {
	switch m {
	case MessageTypeInvalid:
		return "INVALID"
	case MessageTypeCommand:
		return "COMMAND"
	case MessageTypeResponse:
		return "RESPONSE"
	case MessageTypeSignal:
		return "SIGNAL"
	case MessageTypeData:
		return "DATA"
	case MessageTypeAny:
		return "ANY"
	default:
		return "UNKNOWN"
	}
}

type MessageScope int32

const (
	MessageScopeUndefined = MessageScope(pb.MessageScope_UNDEFINED)
	MessageScopeProcess   = MessageScope(pb.MessageScope_PROCESS)
	MessageScopeDevice    = MessageScope(pb.MessageScope_DEVICE)
	MessageScopeNetwork   = MessageScope(pb.MessageScope_NETWORK)
	MessageScopeGlobal    = MessageScope(pb.MessageScope_GLOBAL)
)

func (s MessageScope) String() string {
	switch s {
	case MessageScopeUndefined:
		return "UNDEFINED"
	case MessageScopeProcess:
		return "PROCESS"
	case MessageScopeDevice:
		return "DEVICE"
	case MessageScopeNetwork:
		return "NETWORK"
	case MessageScopeGlobal:
		return "GLOBAL"
	default:
		return "UNDEFINED"
	}
}

const (
	SignalLevelError   = "ERROR"
	SignalLevelWarning = "WARNING"
	SignalLevelInfo    = "INFO"
	SignalLevelDebug   = "DEBUG"
)

type Message interface {
	GetHeader() MessageHeader
	GetTraces() []MessageTrace
	AddTrace(relay string)
}

type MessageHeader interface {
	Hash() [32]byte

	GetTimestamp() time.Time
	GetTimestampString() string
	SetTimestampNow()
	SetTimestamp(t float64)

	GetId() string
	SetId(id string)
	SetIdRandom()

	GetTaskId() string
	SetTaskId(id string)
	SetTaskIdRandom()

	GetTopic() string
	GetScope() MessageScope
	GetName() string
	GetType() MessageType
	GetTypeString() string
	GetOrigin() string
	GetVia() string
	SetVia(via string)
}

type MessageTrace interface {
	GetRelay() string
	GetTime() (time.Time, error)
	GetTimeString() string
}

type DataMessage interface {
	GetHeader() MessageHeader
	GetTraces() []MessageTrace
	AddTrace(relay string)
	GetFormat() DataFormat
	GetRows() []DataRow
	CountRows() int
	AddRow(row DataRow)
	AddRows(rows []DataRow)
}

type DataRow interface {
	GetTimeString() string
	GetFormat() DataFormat
	GetFieldsRaw() map[string]string
	GetTime() time.Time
	SetTime(timestamp time.Time)
	GetFieldInt(field string) *int
	GetFieldIntArray(field string) []int
	GetFieldLong(field string) *int64
	GetFieldUint32(field string) *uint32
	GetFieldUint32Array(field string) []uint32
	GetFieldUint64(field string) *uint64
	GetFieldUint64Array(field string) []uint64
	GetFieldDouble(field string) *float64
	GetFieldDoubleArray(field string) []float64
	GetFieldBool(field string) *bool
	GetFieldString(field string) *string
	GetFieldBinary(field string) []byte
	GetFieldDatetime(s string) *time.Time
	IsFieldNull(field string) bool
	GetAgeMilliseconds() int64
	CountBytes() int
}

type DataFormat interface {
	GetFields() []DataField
	GetFormatId() string
	GetScope() MessageScope
	SetScope(scope MessageScope)
	NewRow(a map[string]interface{}) (DataRow, error)
	NewRowFromPreStringifiedFieldValues(items map[string]string) (DataRow, error)
	GetDataQueueMaxAge() time.Duration
	SetDataQueueMaxAge(duration time.Duration)
	GetDataQueueMaxCount() int
	SetDataQueueMaxCount(length int)
}

type DataField interface {
	GetName() string
	GetType() string
	GetPos() int
}

type SignalMessage interface {
	GetHeader() MessageHeader
	GetTraces() []MessageTrace
	AddTrace(relay string)
	GetType() string
	GetName() string
	GetContent() string
}

type CommandMessage interface {
	GetHeader() MessageHeader
	GetTraces() []MessageTrace
	SetTraces(traces []MessageTrace)
	AddTrace(relay string)
	GetCommand() string
	GetParameters() []map[string]string
	AddParameters(parameters map[string]string)
	SetParameters(parameters []map[string]string)
}

type CommandBody interface {
	GetCommand() string
	GetParameters() []map[string]string
}

type ResponseMessage interface {
	GetHeader() MessageHeader
	GetTraces() []MessageTrace
	SetTraces(traces []MessageTrace)
	AddTrace(relay string)
	GetCommandId() string
	AddResult(results RpcResult)
	SetResults(results []RpcResult)
	CountResults() int
	GetResult(i int) (RpcResultReadonly, error)
	GetResults() []RpcResultReadonly
	HasErrors() bool
	GetErrors() []string
}
