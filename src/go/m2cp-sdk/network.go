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
	// Whether messages shall be acked by or after the message callback.
	// Default: false, defaulting to noAck behavior of [import/github.com/rabbitmq/amqp091-go/Channel.Consume]
	// where messages are automatically acked when sent from the message-hub.
	ManualAck bool
}

type NetworkConnection interface {
	NewNode(name string) (Node, error)

	// SubscribeSignals subscribes to signal messages on the given topics.
	// Acknowledgment is handled automatically. See [NetworkConnection.SubscribeSignalsAckWithOptions] when manual ack control is needed.
	SubscribeSignals(topics []string, handler SignalHandler) error
	// SubscribeSignalsWithOptions is like [NetworkConnection.SubscribeSignals] but accepts additional [SubscriptionOptions].
	SubscribeSignalsWithOptions(topics []string, handler SignalHandler, options SubscriptionOptions) error
	// SubscribeSignalsAckWithOptions is like [NetworkConnection.SubscribeSignalsWithOptions] but with manual acknowledgment control.
	// The handler receives each message together with an [Acknowledger] to explicitly Ack, Nack, or Reject it.
	// Set [SubscriptionOptions.ManualAck] = true to take ownership of acknowledgment; otherwise ack is still automatic and the Acknowledger should not be used.
	// Returns a Subscriber for potential manual unsubscription and queue removal, or an error
	SubscribeSignalsAckWithOptions(topics []string, handler SignalAckHandler, options SubscriptionOptions) (Subscriber, error)

	// SubscribeData subscribes to data messages on the given topics.
	// Acknowledgment is handled automatically. See [NetworkConnection.SubscribeDataAckWithOptions] when manual ack control is needed.
	SubscribeData(topics []string, handler DataHandler) error
	// SubscribeDataWithOptions is like [NetworkConnection.SubscribeData] but accepts additional [SubscriptionOptions].
	SubscribeDataWithOptions(topics []string, handler DataHandler, options SubscriptionOptions) error
	// SubscribeDataAckWithOptions is like [NetworkConnection.SubscribeDataWithOptions] but with manual acknowledgment control.
	// The handler receives each message together with an [Acknowledger] to explicitly Ack, Nack, or Reject it.
	// Set [SubscriptionOptions.ManualAck] = true to take ownership of acknowledgment; otherwise ack is still automatic and the Acknowledger should not be used.
	// Returns a Subscriber for potential manual unsubscription and queue removal, or an error
	SubscribeDataAckWithOptions(topics []string, handler DataAckHandler, options SubscriptionOptions) (Subscriber, error)

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
	//  - m2cp.SignalLevelDebug
	//  - m2cp.SignalLevelInfo
	//  - m2cp.SignalLevelWarning
	//  - m2cp.SignalLevelError
	EmitSignal(name string, value string, logLevel string) error

	// EmitSignalBroadcast sends a signal to all nodes in the network on the broadcast channel. The broadcast channel is reserved for
	// messages of system-wide importance, not application specific messages.
	//  - m2cp.SignalLevelDebug
	//  - m2cp.SignalLevelInfo
	//  - m2cp.SignalLevelWarning
	//  - m2cp.SignalLevelError
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

// Acknowledger provides methods to acknowledge, reject, or requeue a message received from the message-hub.
// It is passed to message callbacks to allow manual acknowledgment control when [SubscriptionOptions.ManualAck] is enabled.
// When ManualAck is false, acknowledgment is handled automatically and the Acknowledger should not be used.
type Acknowledger interface {
	// Ack acknowledges the message, signaling successful processing.
	Ack() error
	// Nack negatively acknowledges the message and requeues it for redelivery.
	Nack() error
	// Reject rejects the message and removes it from the queue without redelivery.
	Reject() error
}

// SignalHandler is the callback type for signal message subscriptions without manual acknowledgment.
// Acknowledgment is handled automatically. Use [SignalAckHandler] when explicit ack control is needed.
type SignalHandler func(message SignalMessage)

// SignalAckHandler is the callback type for signal message subscriptions with manual acknowledgment control.
// The [Acknowledger] must be used to Ack, Nack, or Reject each message iff [SubscriptionOptions.ManualAck] is true.
// Use with [NetworkConnection.SubscribeSignalsAckWithOptions].
type SignalAckHandler func(message SignalMessage, ack Acknowledger)

// DataHandler is the callback type for data message subscriptions without manual acknowledgment.
// Acknowledgment is handled automatically. Use [DataAckHandler] when explicit ack control is needed.
type DataHandler func(message DataMessage)

// DataAckHandler is the callback type for data message subscriptions with manual acknowledgment control.
// The [Acknowledger] must be used to Ack, Nack, or Reject each message iff [SubscriptionOptions.ManualAck] is true.
// Use with [NetworkConnection.SubscribeDataAckWithOptions].
type DataAckHandler func(message DataMessage, ack Acknowledger)

type Sender interface {
	Send(m Message) error
	Stop()
}

type Subscriber interface {
	Stop() // Stop the subscription and release resources. After Stop, the handler will not be called anymore.
	// Remove the queue. This is usually not needed. It is an escape hatch to purposefully destroy a persistent queue.
	// Stop should be called first.
	// Returns true on success, false if the queue was not removed.
	Remove() bool
}
