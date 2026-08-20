package amqp

import (
	"errors"
	"fmt"
	"m2cp"
	"m2cp/messages"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	DebugSubscriber bool = false
)

// Ensure subscriber complies with the exposed m2cp.Subscriber interface
var _ = m2cp.Subscriber(&subscriber{})

type subscriber struct {
	amqpClient           *amqpClient
	ctp                  m2cp.ContextPlus
	domain               m2cp.Domain
	conn                 *amqp.Connection
	amqpChannel          *amqp.Channel
	queue                amqp.Queue
	queueName            string
	persistent           bool
	shuttingDown         bool
	closed               bool
	topics               []string
	messageType          m2cp.MessageType
	callbackSignal       func(msg m2cp.SignalMessage, ack m2cp.Acknowledger)
	callbackData         func(msg m2cp.DataMessage, ack m2cp.Acknowledger)
	callbackCommand      func(msg m2cp.CommandMessage)  // no Acknowledger, as workRpc does ack handling already
	callbackResponse     func(msg m2cp.ResponseMessage) // no Acknowledger, as workRpc does ack handling already
	callbackRaw          func(raw []byte, ack m2cp.Acknowledger)
	running              bool
	waitGroupInit        *sync.WaitGroup
	waitGroupShutdown    *sync.WaitGroup
	options              *m2cp.SubscriptionOptions
	connectedAtLeastOnce bool
}

// acknowledger is an [amqp] specific [m2cp.Acknowledger] implementation that wraps an [amqp.Delivery] to provide Ack, Nack, and Reject functionality.
type acknowledger struct {
	delivery amqp.Delivery
}

// Ack acknowledges the message, signaling successful processing.
// see [amqp.Delivery.Ack] for details
func (cb *acknowledger) Ack() error {
	return cb.delivery.Ack(false)
}

// Nack negatively acknowledges the message and requeues it for redelivery.
// See [amqp.Delivery.Nack] for details
func (cb *acknowledger) Nack() error {
	return cb.delivery.Nack(false, true)
}

// Reject rejects the message and removes it from the queue without redelivery.
// See [amqp.Delivery.Reject] for details
func (cb *acknowledger) Reject() error {
	return cb.delivery.Reject(false)
}

// NewSubscriberSignalsAck creates a new subscriber for signal messages with manual acknowledgment control.
// The callback receives each signal message along with a [m2cp.Acknowledger] to control message disposition.
// Use this with [m2cp.SubscriptionOptions.ManualAck] = true when explicit control over Ack, Nack, or Reject is needed.
// The Acknowledger must not be used when ManualAck is set to false.
//
// Parameters: options configures subscription behavior; topics specifies which topics to subscribe to;
// callbackSignal is called with each received signal message and an Acknowledger; waitGroupInit and waitGroupShutdown coordinate lifecycle.
// Returns a Subscriber or an error if initialization fails.
func NewSubscriberSignalsAck(options m2cp.SubscriptionOptions, topics []string, callbackSignal func(msg m2cp.SignalMessage, ack m2cp.Acknowledger),
	waitGroupInit, waitGroupShutdown *sync.WaitGroup) (m2cp.Subscriber, error) {
	s, err := newSubscriber(&options, topics, waitGroupInit, waitGroupShutdown)
	if err != nil {
		return nil, err
	}
	s.messageType = m2cp.MessageTypeSignal
	s.callbackSignal = callbackSignal
	go s.workerThread()
	return s, nil // s.waitUntilRunning()
}

// NewSubscriberDataAck creates a new subscriber for data messages with manual acknowledgment control.
// The callback receives each data message along with a [m2cp.Acknowledger] to control message disposition.
// Use this with [m2cp.SubscriptionOptions.ManualAck] = true when explicit control over Ack, Nack, or Reject is needed.
// The Acknowledger must not be used when ManualAck is set to false.
//
// Parameters: options configures subscription behavior; topics specifies which topics to subscribe to;
// callbackData is called with each received data message and an Acknowledger; waitGroupInit and waitGroupShutdown coordinate lifecycle.
// Returns a Subscriber or an error if initialization fails.
func NewSubscriberDataAck(options m2cp.SubscriptionOptions, topics []string, callbackData func(msg m2cp.DataMessage, ack m2cp.Acknowledger), waitGroupInit, waitGroupShutdown *sync.WaitGroup) (m2cp.Subscriber, error) {
	s, err := newSubscriber(&options, topics, waitGroupInit, waitGroupShutdown)
	if err != nil {
		return nil, err
	}
	s.messageType = m2cp.MessageTypeData
	s.callbackData = callbackData
	go s.workerThread()
	return s, nil // s.waitUntilRunning()
}

// NewSubscriberCommands creates a new subscriber for RPC command messages with automatic acknowledgment.
// Commands are automatically acknowledged as they are processed through the RPC handler.
//
// Parameters: options configures subscription behavior; domain specifies the receiving domain;
// callbackCommand is called with each received command message; waitGroupInit and waitGroupShutdown coordinate lifecycle.
// Returns a Subscriber or an error if initialization fails.
func NewSubscriberCommands(options m2cp.SubscriptionOptions, domain m2cp.Domain, callbackCommand func(msg m2cp.CommandMessage),
	waitGroupInit, waitGroupShutdown *sync.WaitGroup) (m2cp.Subscriber, error) {
	topics := []string{fmt.Sprintf("command/*.%s", domain.GetName())}
	s, err := newSubscriber(&options, topics, waitGroupInit, waitGroupShutdown)
	if err != nil {
		return nil, err
	}
	s.domain = domain
	s.messageType = m2cp.MessageTypeCommand
	s.callbackCommand = callbackCommand
	go s.workerThread()
	return s, nil // s.waitUntilRunning()
}

// NewSubscriberResponses creates a new subscriber for RPC response messages with automatic acknowledgment.
// Responses are automatically acknowledged as they are processed through the RPC handler.
//
// Parameters: options configures subscription behavior; domain specifies the receiving domain;
// callbackResponse is called with each received response message; waitGroupInit and waitGroupShutdown coordinate lifecycle.
// Returns a Subscriber or an error if initialization fails.
func NewSubscriberResponses(options m2cp.SubscriptionOptions, domain m2cp.Domain, callbackResponse func(msg m2cp.ResponseMessage),
	waitGroupInit, waitGroupShutdown *sync.WaitGroup) (m2cp.Subscriber, error) {
	topics := []string{fmt.Sprintf("response/*.%s", domain.GetName())}
	s, err := newSubscriber(&options, topics, waitGroupInit, waitGroupShutdown)
	if err != nil {
		return nil, err
	}
	s.domain = domain
	s.messageType = m2cp.MessageTypeResponse
	s.callbackResponse = callbackResponse
	go s.workerThread()
	return s, nil // s.waitUntilRunning()
}

// NewSubscriberRawAck creates a new subscriber for raw message bytes with manual acknowledgment control.
// The callback receives each message as raw bytes along with a [m2cp.Acknowledger] to control message disposition.
// Use [m2cp.SubscriptionOptions.ManualAck] = true when explicit control over Ack, Nack, or Reject is needed.
// The Acknowledger must not be used when ManualAck is set to false.
//
// Parameters: options configures subscription behavior; topics specifies which topics to subscribe to;
// callbackRaw is called with each received message as raw bytes and an Acknowledger; waitGroupInit and waitGroupShutdown coordinate lifecycle.
// Returns a Subscriber or an error if initialization fails.
func NewSubscriberRawAck(options m2cp.SubscriptionOptions, topics []string, callbackRaw func(raw []byte, ack m2cp.Acknowledger),
	waitGroupInit, waitGroupShutdown *sync.WaitGroup) (m2cp.Subscriber, error) {
	s, err := newSubscriber(&options, topics, waitGroupInit, waitGroupShutdown)
	if err != nil {
		return nil, err
	}
	s.messageType = m2cp.MessageTypeRaw
	s.callbackRaw = callbackRaw
	go s.workerThread()
	return s, nil // s.waitUntilRunning()
}

func newSubscriber(options *m2cp.SubscriptionOptions, topics []string, waitGroupInit, waitGroupShutdown *sync.WaitGroup) (*subscriber, error) {
	if options == nil {
		return nil, fmt.Errorf("options are required")
	}
	if options.Context == nil {
		return nil, fmt.Errorf("SubscriptionOptions.Context is required")
	}

	s := &subscriber{
		amqpClient:        nil,
		ctp:               options.Context.BranchWithName("networks.amqp.Subscriber"),
		shuttingDown:      false,
		topics:            topics,
		waitGroupInit:     waitGroupInit,
		waitGroupShutdown: waitGroupShutdown,
		options:           options,
	}

	s.waitGroupInit.Add(1)

	if options.PersistenceId != nil {
		s.queueName = *options.PersistenceId
		s.persistent = true
	}

	if DebugSubscriber {
		s.ctp.LogDebug("new amqp %s-subscriber, persistent=%v, q-name=%v", s.messageType, s.persistent, s.queueName)
	}

	return s, nil
}

func (o *subscriber) Stop() {
	o.ctp.Cancel()
	for !o.closed {
		o.ctp.Sleep(10 * time.Millisecond)
	}
}

func (o *subscriber) Remove() bool {
	if o.persistent {
		if err := o.deleteQueue(); err != nil {
			o.ctp.LogError("failed to delete queue '%s': %s", o.queueName, err.Error())
			return false
		}
		o.ctp.LogInfo("deleted queue '%s'", o.queueName)
		return true
	}
	return false
}

func (s *subscriber) doneInit() {
	if !s.connectedAtLeastOnce {
		s.waitGroupInit.Done()
		s.connectedAtLeastOnce = true
	}
}

func (o *subscriber) workerThread() {
	var err error

	o.waitGroupShutdown.Add(1)
	defer func() {
		o.doneInit()
		o.closed = true
		o.running = false
		o.waitGroupShutdown.Done()
		o.ctp.LogDebug("done: amqp %v-subscriber closed", o.messageType)
	}()

	// init amqp connection and setup exchanges
	for {
		if o.amqpClient, err = newAmqpConnection(o.ctp); err == nil {
			break // initialization successful
		}
		o.ctp.Sleep(1 * time.Second)
		if o.ctp.IsCancelled() {
			return
		}
	}

	// main loop
	for {
		switch o.messageType {
		case m2cp.MessageTypeCommand:
			err = o.workRpc("command")
		case m2cp.MessageTypeResponse:
			err = o.workRpc("response")
		default:
			err = o.workMessage()
		}

		restart := !o.ctp.IsCancelled()

		if err != nil {
			o.ctp.LogError("amqp %v-subscriber stopped working, restart=%v, error: %s", o.messageType, restart, err.Error())
		} else if restart {
			o.ctp.LogWarn("amqp %v-subscriber stopped working, restart=%v", o.messageType, restart)
		}

		if !restart {
			o.ctp.LogDebug("amqp %v-subscriber shutting down", o.messageType)
			return
		}

		o.ctp.Sleep(1 * time.Second)
	}
}

func (o *subscriber) activateChannel() error {
	var err error

	if o.conn == nil || o.conn.IsClosed() {
		if o.conn, err = o.amqpClient.getAmqpConnection(); err != nil {
			return fmt.Errorf("failed to connect to AMQP: %w", err)
		}
	}
	if o.amqpChannel == nil || o.amqpChannel.IsClosed() {
		if o.amqpChannel, err = o.conn.Channel(); err != nil {
			return fmt.Errorf("failed to open a channel: %w", err)
		}
	}
	return nil
}

func (o *subscriber) declareQueue() error {
	var err error

	if err = o.activateChannel(); err != nil {
		return err
	}

	var args amqp.Table
	if o.persistent {
		ttl := 48 * time.Hour
		args = amqp.Table{
			"x-expires": int32(ttl.Milliseconds()), // TTL in milliseconds for the queue
		}
	}

	if DebugSubscriber {
		o.ctp.LogDebug("configure q: '%s', persistent=%v, args=%v", o.queueName, o.persistent, args)
	}

	// persistent queues are expected to already exist (unless the app is using it for the very first time on a new device)
	if o.persistent {
		if o.queue, err = o.amqpChannel.QueueDeclarePassive(
			o.queueName,   // name
			o.persistent,  // durable
			!o.persistent, // auto-delete (means: delete the queue when there are no more consumers)
			false,         // exclusive (means: temporary for this connection)
			false,
			args,
		); err != nil {
			o.ctp.LogWarn("queue '%s' does not exist, creating it - this should only happen when this app runs for the very first time", o.queueName)
		} else {
			if DebugSubscriber {
				o.ctp.LogDebug("found preexisting q: persistent q '%s' found on AMQP server", o.queueName)
			}
			return nil
		}
	}

	// create a new queue
	if err = o.activateChannel(); err != nil {
		return err
	}
	if o.queue, err = o.amqpChannel.QueueDeclare(
		o.queueName,   // name
		o.persistent,  // durable
		!o.persistent, // auto-delete (means: delete the queue when there are no more consumers)
		false,         // exclusive (means: temporary for this connection)
		false,
		args,
	); err != nil {
		return err
	}
	if DebugSubscriber {
		o.ctp.LogDebug("created new q:  '%s'", o.queueName)
	}
	return nil
}

func (o *subscriber) deleteQueue() error {
	var err error
	if err = o.activateChannel(); err != nil {
		return err
	}
	if _, err = o.amqpChannel.QueueDelete(o.queueName, false, false, false); err != nil {
		return err
	}
	return nil
}

func (o *subscriber) connInit() error {
	var err error
	if err = o.declareQueue(); err == nil {
		o.doneInit()
		return nil
	}
	var amqpErr *amqp.Error
	if errors.As(err, &amqpErr) {
		if amqpErr.Code == 406 {
			// 406 is the code for "PRECONDITION_FAILED"
			// this means the queue already exists and the parameters are different - we'll have to delete it and retry
			o.ctp.LogWarn("existing queue mismatches parameters - deleting and re-declaring '%s'", o.queueName)
			if err = o.deleteQueue(); err != nil {
				return fmt.Errorf("failed to delete queue '%s': %w", o.queueName, err)
			}
			if err = o.declareQueue(); err != nil {
				return fmt.Errorf("failed to re-declare queue '%s': %w", o.queueName, err)
			}
			o.ctp.LogDebug("queue '%s' re-declared with new parameters", o.queueName)
			o.doneInit()
			return nil
		} else {
			return fmt.Errorf("failed to declare queue '%s'. AMQP Error: Code: %d, Reason: %s", o.queueName, amqpErr.Code, amqpErr.Reason)
		}
	}
	return fmt.Errorf("failed to declare queue '%s': %w", o.queueName, err)
}

func (o *subscriber) connCleanup() {
	o.running = false
	if o.amqpChannel != nil {
		_ = o.amqpChannel.Close()
	}
	if o.conn != nil {
		_ = o.conn.Close()
	}
}

func (o *subscriber) workRpc(role string) error {
	defer o.connCleanup()
	var err error

	o.ctp.LogDebug("lstn: %s", o.queue.Name)
	o.queueName = fmt.Sprintf("%s.%s", role, o.domain.GetName())
	if err = o.connInit(); err != nil {
		return err
	}
	if err = o.amqpChannel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}
	var deliveryChan <-chan amqp.Delivery
	if deliveryChan, err = o.amqpChannel.Consume(
		o.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	); err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}
	errorChan := o.amqpChannel.NotifyClose(make(chan *amqp.Error))

	o.running = true
	for o.running {
		select {
		case d, ok := <-deliveryChan:
			if !ok {
				return fmt.Errorf("delivery channel is closed")
			}
			o.ctp.LogDebug("recv: %s", d.RoutingKey)
			if err = d.Ack(false); err != nil {
				return fmt.Errorf("failed ack for %s message: %w", role, err)
			}
			if role == "command" {
				go o.handleCommandMessage(d.Body, d.ReplyTo)
			} else if role == "response" {
				go o.handleResponseMessage(d.Body, d.ReplyTo)
			}
		case d := <-errorChan:
			o.running = false
			return fmt.Errorf("%s channel closed: %s", role, d.Error())
		case <-o.ctp.Done():
			return nil
		}
	}
	return nil
}

func (o *subscriber) handleCommandMessage(msg []byte, replyTo string) {
	defer func() {
		if r := recover(); r != nil {
			o.ctp.LogError("panic recovered in command execution: %v", r)
		}
	}()
	var cmd m2cp.CommandMessage
	var err error
	if cmd, err = messages.CommandFromBinary(msg); err != nil {
		o.ctp.LogError("failed parsing incoming command message: %s", err.Error())
		return
	}
	cmd.GetHeader().SetVia(replyTo)
	o.callbackCommand(cmd)
}

func (o *subscriber) handleResponseMessage(msg []byte, replyTo string) {
	defer func() {
		if r := recover(); r != nil {
			o.ctp.LogError("panic recovered in response callback: %v", r)
		}
	}()
	var resp m2cp.ResponseMessage
	var err error
	if resp, err = messages.ResponseFromBinary(msg); err != nil {
		o.ctp.LogError("failed parsing incoming response message: %s", err.Error())
		return
	}
	o.callbackResponse(resp)
}

func (o *subscriber) workMessage() (err error) {
	defer o.connCleanup()
	if err = o.connInit(); err != nil {
		return err
	}

	for _, topic := range o.topics {
		routingKey := topic2RoutingKey(topic)
		exchangeName, _ := messageType2Exchange(o.messageType)
		err = o.amqpChannel.QueueBind(o.queue.Name, routingKey, exchangeName, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind a queue: %w", err)
		}
		o.ctp.LogDebug("subs: %s@%s(%s)", routingKey, exchangeName, o.queue.Name)
	}

	msgs, err := o.amqpChannel.Consume(o.queue.Name, "", !o.options.ManualAck, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}
	errorChan := o.amqpChannel.NotifyClose(make(chan *amqp.Error))

	go func() {
		for msg := range msgs {
			err2 := o.handleIncoming(msg)
			if err2 != nil {
				o.ctp.LogError(err2.Error())
			}
		}
	}()
	o.ctp.LogDebug("lstn: %s", o.queue.Name)
	o.running = true
	for o.running {
		// as soon as context is canceled, we should stop the worker
		select {
		case d := <-errorChan:
			o.running = false
			return fmt.Errorf("subs queue %s closed: %s", o.queue.Name, d.Error())
		case <-o.ctp.Done():
			return nil
		}
	}
	return nil
}

func (o *subscriber) handleIncoming(msg amqp.Delivery) (errResult error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			errResult = fmt.Errorf("recovered from panic in callback: %v", r)
		}
	}()
	// for subscribers, who want to receive the raw bytes of the message
	if o.callbackRaw != nil {
		o.callbackRaw(msg.Body, &acknowledger{msg})
	}
	if o.callbackSignal == nil && o.callbackData == nil {
		return
	}
	mType := o.typeFromAmqp(msg)
	switch mType {
	case m2cp.MessageTypeSignal:
		if o.callbackSignal != nil {
			var signalMsg m2cp.SignalMessage
			signalMsg, err = messages.SignalFromBinary(msg.Body)
			if err != nil {
				return fmt.Errorf("failed to parse signal message: %w", err)
			}
			o.callbackSignal(signalMsg, &acknowledger{msg})
		}
	case m2cp.MessageTypeData:
		if o.callbackData != nil {
			var dataMsg m2cp.DataMessage
			dataMsg, err = messages.DataFromBinary(msg.Body)
			if err != nil {
				return fmt.Errorf("failed to parse data message: %w", err)
			}
			o.callbackData(dataMsg, &acknowledger{msg})
		}
	default:
		return fmt.Errorf("message of unexpected type: %s", msg.Body)
	}
	return nil
}

func (o *subscriber) typeFromAmqp(msg amqp.Delivery) m2cp.MessageType {
	switch msg.RoutingKey[0:5] {
	case "data.":
		return m2cp.MessageTypeData
	case "signa":
		return m2cp.MessageTypeSignal
	case "comma":
		return m2cp.MessageTypeCommand
	case "respo":
		return m2cp.MessageTypeResponse
	default:
		return m2cp.MessageTypeInvalid
	}
}

func (o *subscriber) waitUntilRunning() error {
	timeStart := time.Now()
	for !o.running {
		o.ctp.Sleep(1 * time.Millisecond)
		if time.Now().Sub(timeStart) > 5*time.Second {
			o.ctp.Cancel()
			return fmt.Errorf("subscriber failed to start")
		}
	}
	return nil
}
