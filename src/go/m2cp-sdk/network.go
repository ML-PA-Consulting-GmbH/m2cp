package m2cp

import (
	"time"
)

type Domain interface {
	GetName() string
	GetDeviceName() string
	GetAppName() string
}

type Address interface {
	GetAddress() string
	GetSubtopic() string
	GetDeviceName() string
	// GetDeviceIp if the device token of an address is IP based, this method returns the IP address.
	GetDeviceIp() string
	// GetDevicePort if the device token of an address is IP based, this method returns the port.
	GetDevicePort() string
	GetAppName() string
	GetNodeName() string
	IsIpRouted() bool
}

type SubscriptionOptions struct {
	Context ContextPlus
	// Set a persistence ID to enable subscription persistence. WARNING: id must be used exclusively by a single instance to avoid name collisions and blocking!
	PersistenceId *string
}

type NetworkConnection interface {
	NewNode(name string) (Node, error)
	SubscribeSignals(topics []string, handler SignalHandler) error
	SubscribeSignalsWithOptions(topics []string, handler SignalHandler, options SubscriptionOptions) error
	SubscribeData(topics []string, handler DataHandler) error
	SubscribeDataWithOptions(topics []string, handler DataHandler, options SubscriptionOptions) error
	SendMessage(message Message) error
	GetDomain() Domain
	GetContext() ContextPlus
	DiscoverNodes() error
	GetKnownNodeAddresses() []string
	AwaitReady()
	Close()
	GetMessageSerializer() MessageSerializer
	GetStats() NetworkConnectionStats
}

type NetworkConnectionStats interface {
	GetSenderStats() SenderStats
	// GetReceiverStats() WorkerStats
}

type SenderStats interface {
	GetTransmitted() uint64
	GetAcked() uint64
	GetErrors() uint64
	GetQueueSendLen() uint64
	GetQueueSendMax() uint64
	GetQueueAckLen() uint64
	GetQueueAckMax() uint64
	GetConnected() bool
}

type NetworkConnectionOptions struct {
	AppName           string // leave empty for auto-configure
	DeviceName        string // leave empty for auto-configure
	SendQueueSize     int
	MessageSerializer MessageSerializer
}

type MessageSerializer byte

const (
	MessageSerializerJson        MessageSerializer = 0
	MessageSerializerJsonDeflate MessageSerializer = 1
	MessageSerializerProtobuf    MessageSerializer = 2

	MessageSerializerDefault MessageSerializer = 0
)

type RemoteProcedureCallOptions struct {
	Trace []MessageTrace
	DTLS  DTLSConfig
}

type Node interface {
	Address
	GetDomain() Domain
	Uptime() int64
	GetName() string
	GetContext() ContextPlus
	SetRpcHandler(handler RpcHandler)
	SetDataQueueMaxAge(duration time.Duration)
	SetDataQueueMaxCount(int)
	EmitDataRow(DataRow) error
	EmitDataRows([]DataRow) error
	GetFailedToSendCount() int

	// EmitSignal emits a signal from a node. For "logLevel" use:
	// m2cp.SignalLevelDebug
	// m2cp.SignalLevelInfo
	// m2cp.SignalLevelWarning
	// m2cp.SignalLevelError
	EmitSignal(name string, value string, logLevel string) error

	// EmitSignalBroadcast sends a signal to all nodes in the network on the broadcast channel. The broadcast channel is reserved for
	// messages of system-wide importance, not application specific messages.
	// m2cp.SignalLevelDebug
	// m2cp.SignalLevelInfo
	// m2cp.SignalLevelWarning
	// m2cp.SignalLevelError
	EmitSignalBroadcast(name string, value string, logLevel string) error

	RemoteProcedureCall(to Address, command string, parameters map[string]string) (<-chan RpcResultReadonly, error)
	RemoteProcedureCalls(to Address, command string, parameters []map[string]string) (<-chan []RpcResultReadonly, error)
	RemoteProcedureCallsWithOptions(to Address, command string, parameters []map[string]string, options RemoteProcedureCallOptions) (<-chan []RpcResultReadonly, error)
	SetRpcTimeout(duration time.Duration)

	//RemoteProcedureCallWithParams(to string, command string, parameters commands.CommandParameters) commands.RpcBuilder
	//RemoteProcedureCallNode(to Node, command string) commands.RpcBuilder
	//RemoteProcedureCallNodeWithParams(to Node, command string, parameters commands.CommandParameters) commands.RpcBuilder
	//RemoteProcedureCallAddress(to messages.IM2CPAddress, command string) commands.RpcBuilder
	//RemoteProcedureCallAddressWithParams(to messages.IM2CPAddress, command string, parameters commands.CommandParameters) commands.RpcBuilder
	//RemoteProcedureCallCommand(command messages.ICommandMessage) commands.RpcBuilder

	//GetConnector() NetworkConnection
	//ExecuteRpc(command messages.ICommandMessage, callback func(response messages.IResponseMessage))
	//ExecuteRpcSync(command messages.ICommandMessage) messages.IResponseMessage

	//GetSendQueueLength() int
}

type SignalHandler func(message SignalMessage)

type DataHandler func(message DataMessage)

type Sender interface {
	Send(m Message) error
	Stop()
}

type Subscriber interface {
	Stop()
}
