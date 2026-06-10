package networks

import (
	"fmt"
	"m2cp"
	"m2cp/messages"
	"m2cp/networks/amqp"
	"m2cp/networks/stats"
	"m2cp/rpc/rpctypes"
	"strings"
	"sync"
	"time"
)

type NetworkConnectionInternal interface {
	m2cp.NetworkConnection
}

var (
	domainLocks map[string]*sync.Mutex = make(map[string]*sync.Mutex)
)

type networkConnectionAmqp struct {
	cfg                    *NetworkContext
	ctp                    m2cp.ContextPlus
	pluginCounter          int32
	timeStart              int64
	knownHosts             knownHosts
	signalBroadcastHandler m2cp.SignalHandler
	sender                 m2cp.Sender
	subscriberCommands     m2cp.Subscriber
	subscriberResponses    m2cp.Subscriber
	waitGroupShutdown      sync.WaitGroup
	waitGroupInit          sync.WaitGroup
	messageSerializer      m2cp.MessageSerializer
	stats                  *stats.NetworkStats
}

func (nc *networkConnectionAmqp) GetStats() m2cp.NetworkConnectionStats {
	return nc.stats
}

func NewNetworkConnectionWithOptions(ctx m2cp.ContextPlus, options m2cp.NetworkConnectionOptions) (m2cp.NetworkConnection, error) {
	if ctx.IsCancelled() {
		return nil, fmt.Errorf("can't start network connection with cancelled context")
	}
	ctx = ctx.Branch()

	d, err := newDomain(ctx, options.AppName, options.DeviceName)
	if err != nil {
		return nil, err
	}

	// only one network connection can exist at a time per domain - otherwise we would have naming conflicts on the amqp exchange
	domainName := d.GetName()
	ctx.LogDebug("acquiring lock for '%s' network connection  (only one network connection can exist at a time per domain)", domainName)
	if _, exists := domainLocks[domainName]; !exists {
		domainLocks[domainName] = &sync.Mutex{}
	}
	domainLocks[domainName].Lock()
	ctx.LogDebug("acquired lock for '%s' - starting up", domainName)

	// release lock on panic during constructor
	defer func() {
		if r := recover(); r != nil {
			domainLocks[domainName].Unlock()
			panic(r)
		}
	}()

	nc, err := newNetworkConnection(ctx, d, options)
	if err != nil {
		domainLocks[domainName].Unlock()
		return nil, err
	}

	// watchdog: release lock when network connection is closed
	go func() {
		defer func() {
			if r := recover(); r != nil {
				domainLocks[domainName].Unlock()
				panic(r)
			}
		}()

		<-ctx.Done()
		ctx.LogDebug("network connection watchdog: context cancelled -> closing network connection and releasing lock")
		nc.Close()
		domainLocks[domainName].Unlock()
	}()

	return nc, nil
}

func NewNetworkConnection(ctx m2cp.ContextPlus) (m2cp.NetworkConnection, error) {
	return NewNetworkConnectionWithOptions(ctx, m2cp.NetworkConnectionOptions{
		MessageSerializer: m2cp.MessageSerializerDefault,
	})
}

func newNetworkConnection(ctpParent m2cp.ContextPlus, domain *domain, options m2cp.NetworkConnectionOptions) (m2cp.NetworkConnection, error) {
	var err error
	cfg := NewNetworkContext(domain, ctpParent)
	conn := networkConnectionAmqp{
		cfg:               cfg,
		ctp:               ctpParent.SetModule(fmt.Sprintf("networks.NetworkConnection[%s]", domain.GetName())),
		knownHosts:        newKnownHosts(),
		messageSerializer: options.MessageSerializer,
		stats:             &stats.NetworkStats{},
	}
	conn.timeStart = time.Now().Unix()

	// not implemented: conn.waitUntilConnected()

	conn.waitGroupInit.Add(1) // wait until all initializations are done - master entry

	if conn.sender, err = amqp.NewSender(
		conn.cfg.ctp,
		&conn.waitGroupInit,
		&conn.waitGroupShutdown,
		options.SendQueueSize,
		conn.messageSerializer,
		conn.stats,
	); err != nil {
		conn.cfg.ctp.Cancel()
		conn.waitGroupInit.Done() // release master entry
		return nil, err
	}

	if err = conn.SubscribeSignals([]string{"sys-broadcast"}, conn.systemBroadCastHandler); err != nil {
		conn.cfg.ctp.Cancel()
		conn.waitGroupInit.Done() // release master entry
		return nil, err
	}

	if conn.subscriberCommands, err = amqp.NewSubscriberCommands(
		m2cp.SubscriptionOptions{Context: conn.ctp}, cfg.Domain, conn.commandHandler, &conn.waitGroupInit, &conn.waitGroupShutdown); err != nil {
		conn.cfg.ctp.Cancel()
		conn.waitGroupInit.Done() // release master entry
		return nil, err
	}

	if conn.subscriberResponses, err = amqp.NewSubscriberResponses(
		m2cp.SubscriptionOptions{Context: conn.ctp}, cfg.Domain, conn.responseHandler, &conn.waitGroupInit, &conn.waitGroupShutdown); err != nil {
		conn.cfg.ctp.Cancel()
		conn.waitGroupInit.Done() // release master entry
		return nil, err
	}

	// NOTE: proactive discovery is deactivated to save bandwidth
	// conn.DiscoverNodes()

	conn.waitGroupInit.Done() // release master entry - now the individual threads have their own entry

	return &conn, nil
}

func (nc *networkConnectionAmqp) GetMessageSerializer() m2cp.MessageSerializer {
	return nc.messageSerializer
}

func (nc *networkConnectionAmqp) AwaitReady() {
	nc.waitGroupInit.Wait()
	nc.ctp.LogDebug("AwaitReady() -> network connection is ready")
}

func (nc *networkConnectionAmqp) NewNode(name string) (m2cp.Node, error) {
	nd := newNode(nc, name)
	nc.cfg.AddNode(nd)
	err := nd.EmitSignalBroadcast("hello!", fmt.Sprintf("%d", nd.Uptime()), m2cp.SignalLevelInfo)
	if err != nil {
		return nil, err
	}
	return nd, nil
}

func (nc *networkConnectionAmqp) Close() {
	nc.ctp.Cancel()
	nc.waitGroupShutdown.Wait()
}

func (nc *networkConnectionAmqp) SendMessage(message m2cp.Message) error {
	return nc.sender.Send(message)
}

func (nc *networkConnectionAmqp) GetContext() m2cp.ContextPlus {
	return nc.cfg.ctp
}

func (nc *networkConnectionAmqp) GetDomain() m2cp.Domain {
	return nc.cfg.Domain
}

func (nc *networkConnectionAmqp) SubscribeData(topics []string, handler m2cp.DataHandler) error {
	return nc.SubscribeDataWithOptions(topics, handler, m2cp.SubscriptionOptions{Context: nc.ctp})
}

func (nc *networkConnectionAmqp) SubscribeDataWithOptions(topics []string, handler m2cp.DataHandler, options m2cp.SubscriptionOptions) error {
	topicsPrefixed := make([]string, len(topics))
	for i, topic := range topics {
		topicsPrefixed[i] = fmt.Sprintf("data/%s", topic)
	}
	_, err := amqp.NewSubscriberData(nc.normalizeSubscriptionOptions(options), topicsPrefixed, handler, &nc.waitGroupInit, &nc.waitGroupShutdown)
	return err
}

func (nc *networkConnectionAmqp) SubscribeSignals(topics []string, handler m2cp.SignalHandler) error {
	return nc.SubscribeSignalsWithOptions(topics, handler, m2cp.SubscriptionOptions{Context: nc.ctp})
}

func (nc *networkConnectionAmqp) SubscribeSignalsWithOptions(topics []string, handler m2cp.SignalHandler, options m2cp.SubscriptionOptions) error {
	topicsPrefixed := make([]string, len(topics))
	for i, topic := range topics {
		topicsPrefixed[i] = fmt.Sprintf("signal/%s", topic)
	}
	_, err := amqp.NewSubscriberSignals(nc.normalizeSubscriptionOptions(options), topicsPrefixed, handler, &nc.waitGroupInit, &nc.waitGroupShutdown)
	return err
}

func (nc *networkConnectionAmqp) GetKnownNodeAddresses() []string {
	// nc.DiscoverNodes() // NOTE: proactive discovery is deactivated to reduce network traffic
	return nc.knownHosts.getKnownHostAddresses()
}

func (nc *networkConnectionAmqp) normalizeSubscriptionOptions(options m2cp.SubscriptionOptions) m2cp.SubscriptionOptions {
	if options.Context == nil {
		options.Context = nc.ctp
	}
	if options.PersistenceId != nil {
		*options.PersistenceId = fmt.Sprintf("%s.%s", *options.PersistenceId, nc.cfg.Domain.GetAppName())
	}
	return options
}

func (nc *networkConnectionAmqp) systemBroadCastHandler(message m2cp.SignalMessage) {
	switch message.GetName() {
	case "hello!":
		nc.ctp.LogDebug("received hello! from %s", message.GetHeader().GetOrigin())
		nc.knownHosts.seen(message.GetHeader().GetOrigin(), message.GetHeader().GetTimestamp())
	case "hello?":
		nc.discoveryReply()
	default:
		nc.ctp.LogError("unknown signal: name=%s, content=%s", message.GetName(), message.GetContent())
	}
}

func (nc *networkConnectionAmqp) discoveryReply() {
	for _, n := range nc.cfg.GetNodes() {
		nc.ctp.LogDebug("reply with hello! (%s)", n.GetName())
		err := n.EmitSignalBroadcast("hello!", fmt.Sprintf("%d", n.Uptime()), m2cp.SignalLevelInfo)
		if err != nil {
			nc.ctp.LogError("failed to emit 'hello!' signal: %s", err.Error())
		}
	}
}

func (nc *networkConnectionAmqp) waitUntilConnected() {
	// TODO: Implementation
}

func (nc *networkConnectionAmqp) DiscoverNodes() error {
	var err error
	var msg m2cp.SignalMessage
	origin := nc.GetDomain().GetName()
	addr, err := messages.NewAddressWithSubtopic(nc.ctp, origin, "sys-broadcast")
	if err != nil {
		return fmt.Errorf("failed to create address for 'hello?' signal: %s", err.Error())
	}
	if msg, err = messages.NewSignalMessage(addr, "hello?", "", m2cp.SignalLevelInfo, m2cp.MessageScopeNetwork); err != nil {
		return fmt.Errorf("failed to create 'hello?' signal: %s", err.Error())
	}
	if err = nc.sender.Send(msg); err != nil {
		return fmt.Errorf("failed to send 'hello?' signal: %s", err.Error())
	}
	return nil
}

func (nc *networkConnectionAmqp) commandHandler(message m2cp.CommandMessage) {
	message.AddTrace(nc.cfg.Domain.GetName())

	topic := message.GetHeader().GetTopic()
	nodeName := topic[8:strings.IndexRune(topic, '.')]

	var err error
	var n m2cp.Node
	n, err = nc.cfg.GetNodeByName(nodeName)

	var response m2cp.ResponseMessage
	if err == nil {
		response, err = n.(*node).handleRpcCall(message)
	}

	if err != nil {
		var errNewResponseMessage error
		response, errNewResponseMessage = messages.NewResponseMessageFailed(message, err.Error())
		if errNewResponseMessage != nil {
			nc.ctp.LogError("failed to create failure response message: %s", errNewResponseMessage.Error())
			return
		}
		response.SetResults([]m2cp.RpcResult{
			rpctypes.NewRpcResultSuccess(err.Error()),
		})
	}

	response.AddTrace(nc.cfg.Domain.GetName())
	err = nc.SendMessage(response)
	if err != nil {
		nc.ctp.LogError("failed to send response message: %s", err.Error())
	}
}

func (nc *networkConnectionAmqp) responseHandler(message m2cp.ResponseMessage) {
	var err error
	message.AddTrace(nc.cfg.Domain.GetName())
	topic := message.GetHeader().GetTopic()
	nodeName := topic[9:strings.IndexRune(topic, '.')]
	var n m2cp.Node
	if n, err = nc.cfg.GetNodeByName(nodeName); err != nil {
		nc.ctp.LogError("responseHandler: node not found: %s", nodeName)
		return
	}
	n.(*node).handleRpcResponse(message)
}
