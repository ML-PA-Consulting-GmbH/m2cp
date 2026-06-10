package networks

import (
	"fmt"
	"m2cp"
	coap_client "m2cp/coap/coap-client"
	"m2cp/messages"
	"m2cp/rpc/rpctypes"
	"sync"
	"time"

	"github.com/plgd-dev/go-coap/v3/message/codes"
)

type node struct {
	name              string
	ctp               m2cp.ContextPlus
	networkConnection m2cp.NetworkConnection
	timeStart         int64
	dataQueues        map[int64]*dataQueue
	dataQueuesLock    sync.Mutex
	dataQueueMaxAge   time.Duration
	dataQueueMaxCount int
	broadcastAddress  m2cp.Address
	commandHandler    m2cp.RpcHandler
	pendingRPCs       map[string]chan []m2cp.RpcResultReadonly
	pendingRPCsLock   sync.Mutex
	rpcTimeout        time.Duration
	messageSerializer m2cp.MessageSerializer
}

type dataQueue struct {
	rows        []m2cp.DataRow
	lastUpdated time.Time
	size        int
	ch          chan m2cp.DataRow
}

func newNode(networkConnection m2cp.NetworkConnection, name string) *node {
	ctx := networkConnection.GetContext()
	ctx.SetModule("networks.Node")
	n := &node{
		name:              name,
		ctp:               ctx,
		networkConnection: networkConnection,
		timeStart:         time.Now().Unix(),
		dataQueues:        make(map[int64]*dataQueue),
		dataQueueMaxAge:   5000 * time.Millisecond,
		dataQueueMaxCount: 100,
		pendingRPCs:       make(map[string]chan []m2cp.RpcResultReadonly),
		rpcTimeout:        5000 * time.Millisecond,
		messageSerializer: networkConnection.GetMessageSerializer(),
	}
	ctp := networkConnection.GetContext()
	var err error
	n.broadcastAddress, err = messages.NewAddressWithSubtopic(ctp, n.GetAddress(), "sys-broadcast")
	if err != nil {
		ctx.LogError("failed to create broadcast address: %s", err.Error())
		return nil
	}
	ctx.LogDebug("created node %s", n.GetAddress())

	//if n.networkConnection.DoNetworkDiscovery {
	//	log.Info().Msg(fmt.Sprintf("[node:%s] say hello..", name))
	//	n.EmitSignalBroadcast("hello!")
	//}
	//n.InitDataQueueAutoSend()
	return n
}

func (n *node) GetName() string {
	return n.name
}

func (n *node) GetContext() m2cp.ContextPlus {
	return n.ctp
}

func (n *node) GetAddress() string {
	return n.name + "." + n.networkConnection.GetDomain().GetName()
}

func (n *node) GetSubtopic() string {
	return n.name + "/"
}

func (n *node) GetDeviceName() string {
	return n.GetDomain().GetDeviceName()
}

// GetDeviceIp nodes are not IP based and thus do not have an IP address. IP addresses only play a role when trying to reach remote devices
// This method is only implemented to fulfill the interface requirements of Node being an Address
func (n *node) GetDeviceIp() string {
	return ""
}

// GetDevicePort nodes are not IP based and thus do not have an IP port. IP addresses only play a role when trying to reach remote devices
// This method is only implemented to fulfill the interface requirements of Node being an Address
func (n *node) GetDevicePort() string {
	return ""
}

func (n *node) GetAppName() string {
	return n.GetDomain().GetAppName()
}

func (n *node) GetNodeName() string {
	return n.name
}

// IsIpRouted tells, if an m2cp-address contains an IP section. As node always returns it's own address as
// derived from the device serial, this is always false.
func (n *node) IsIpRouted() bool {
	return false
}

func (n *node) GetDomain() m2cp.Domain {
	return n.networkConnection.GetDomain()
}

func (n *node) Uptime() int64 {
	return time.Now().Unix() - n.timeStart
}

func (n *node) SetRpcHandler(handler m2cp.RpcHandler) {
	n.commandHandler = handler
}

func (n *node) SetDataQueueMaxAge(duration time.Duration) {
	n.dataQueueMaxAge = duration
}

func (n *node) SetDataQueueMaxCount(length int) {
	n.dataQueueMaxCount = length
}

func (n *node) EmitDataRow(data m2cp.DataRow) error {
	format := data.GetFormat()
	id := messages.GetInternalFormatId(format)

	// create data queue if necessary
	n.dataQueuesLock.Lock()
	queue, ok := n.dataQueues[id]
	if !ok {
		queue = &dataQueue{
			rows:        []m2cp.DataRow{},
			lastUpdated: time.Now(),
			size:        0,
			ch:          make(chan m2cp.DataRow, 100),
		}
		n.dataQueues[id] = queue
		go n.dataQueueManage(format, queue)
	}
	n.dataQueuesLock.Unlock()

	// push data to queue
	queue.ch <- data

	// we can't return an error here, as the queue is managed in the background
	// and we can't know if it will ever be processed
	return nil
}

func (n *node) EmitDataRows(data []m2cp.DataRow) error {
	var err error
	for _, row := range data {
		err = n.EmitDataRow(row)
		if err != nil {
			return err
		}
	}
	return nil
}

// EmitSignal emits a signal with global scope, thus being sent to the cloud - mainly used for logging, reporting system status, etc.
func (n *node) EmitSignal(name, value, level string) error {
	return n.emitSignalScoped(name, value, level, m2cp.MessageScopeGlobal, false)
}

// EmitSignalBroadcast sends a signal to all nodes in the network on the broadcast channel. The broadcast channel is reserved for
// messages of system-wide importance, not application specific messages.
func (n *node) EmitSignalBroadcast(name, value, level string) error {
	msg, err := messages.NewSignalMessage(n.broadcastAddress, name, value, level, m2cp.MessageScopeNetwork)
	if err != nil {
		return err
	}
	return n.networkConnection.SendMessage(msg)
}

// EmitSignalProcess sends a signal to all threads within the same application process. Useful for inter-thread coordination
func (n *node) EmitSignalProcess(name, value, level string) error {
	return n.emitSignalScoped(name, value, level, m2cp.MessageScopeProcess, false)
}

// EmitSignalDevice sends a signal to all applications running on the same device. Useful for inter-application coordination
func (n *node) EmitSignalDevice(name, value, level string) error {
	return n.emitSignalScoped(name, value, level, m2cp.MessageScopeDevice, false)
}

// EmitSignalNetwork sends a signal to all devices in the same network. Useful for network-wide coordination
func (n *node) EmitSignalNetwork(name, value, level string) error {
	return n.emitSignalScoped(name, value, level, m2cp.MessageScopeNetwork, false)
}

func (n *node) emitSignalScoped(name, value, level string, scope m2cp.MessageScope, broadcast bool) error {
	var addr m2cp.Address
	if broadcast {
		addr = n.broadcastAddress
	} else {
		addr = n
	}
	msg, err := messages.NewSignalMessage(addr, name, value, level, scope)
	if err != nil {
		return err
	}
	return n.networkConnection.SendMessage(msg)
}

func (n *node) RemoteProcedureCall(to m2cp.Address, command string, parameters map[string]string) (<-chan m2cp.RpcResultReadonly, error) {
	res, err := n.RemoteProcedureCallsWithOptions(to, command, []map[string]string{parameters}, m2cp.RemoteProcedureCallOptions{})
	if err != nil {
		return nil, err
	}
	resultCh := make(chan m2cp.RpcResultReadonly)
	go func() {
		for results := range res {
			resultCh <- results[0]
			break
		}
		close(resultCh)
	}()
	return resultCh, nil
}

func (n *node) RemoteProcedureCallSync(to m2cp.Address, command string, parameters map[string]string) (m2cp.RpcResultReadonly, error) {
	res, err := n.RemoteProcedureCall(to, command, parameters)
	if err != nil {
		return nil, err
	}
	results := []m2cp.RpcResultReadonly{}
	for result := range res {
		results = append(results, result)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no results received")
	}
	if len(results) > 1 {
		return nil, fmt.Errorf("more than one result received")
	}
	return results[0], nil
}

func (n *node) RemoteProcedureCalls(to m2cp.Address, command string, parameters []map[string]string) (<-chan []m2cp.RpcResultReadonly, error) {
	return n.RemoteProcedureCallsWithOptions(to, command, parameters, m2cp.RemoteProcedureCallOptions{})
}

func (n *node) RemoteProcedureCallsWithOptions(to m2cp.Address, command string, parameters []map[string]string, options m2cp.RemoteProcedureCallOptions) (<-chan []m2cp.RpcResultReadonly, error) {
	if to == nil {
		return nil, fmt.Errorf("to address is nil")
	}
	cmdMsg, err := messages.NewCommandMessage(n, to, command)
	if err != nil {
		return nil, err
	}
	if options.Trace != nil {
		cmdMsg.SetTraces(options.Trace)
	}
	cmdMsg.AddTrace(n.GetAddress())

	if parameters != nil {
		cmdMsg.SetParameters(parameters)
	}
	cmdMsgId := cmdMsg.GetHeader().GetId()

	// Channel for handing over response to the caller
	resultCh := make(chan []m2cp.RpcResultReadonly, 1)

	if to.IsIpRouted() {
		// Send via CoAP
		var cmdMsgBytes []byte
		if cmdMsgBytes, err = messages.ToBinary(cmdMsg, n.messageSerializer); err != nil {
			return nil, err
		}

		go func() {
			defer func() {
				close(resultCh)
			}()

			host := to.GetDeviceName()
			const rpcCoapEndpoint = "/m2cp/rpc"

			var coapClient m2cp.CoapClient
			if options.DTLS != nil {
				coapClient, err = coap_client.NewClientWithOptions(n.ctp, host, m2cp.CoapClientOptions{DTLS: options.DTLS})
				if err != nil {
					resultCh <- []m2cp.RpcResultReadonly{
						rpctypes.NewRpcResultError(fmt.Sprintf("failed creating CoAP DTLS client: %s", err), m2cp.RpcErrorCodeFailed),
					}
					return
				}
			} else {
				coapClient, err = coap_client.NewClient(n.ctp, host)
				if err != nil {
					resultCh <- []m2cp.RpcResultReadonly{
						rpctypes.NewRpcResultError(fmt.Sprintf("failed creating CoAP client: %s", err), m2cp.RpcErrorCodeFailed),
					}
					return
				}
			}
			coapClient.SetRequestTimeout(n.rpcTimeout)

			cmdMsg.AddTrace(fmt.Sprintf("POST CoAP://%s%s", host, rpcCoapEndpoint))

			var coapRes m2cp.CoapResponse
			if coapRes, err = coapClient.Post(rpcCoapEndpoint, "", cmdMsgBytes); err != nil {
				resultCh <- []m2cp.RpcResultReadonly{
					rpctypes.NewRpcResultError(err.Error(), m2cp.RpcErrorCodeFailed),
				}
				return
			}
			if coapRes.GetResponseCode() != codes.Content {
				resultCh <- []m2cp.RpcResultReadonly{
					rpctypes.NewRpcResultError(fmt.Sprintf("bad CoAP response code %d", coapRes.GetResponseCode()), m2cp.RpcErrorCodeFailed),
				}
				return
			}
			var resMsg m2cp.ResponseMessage
			if resMsg, err = messages.ResponseFromBinary(coapRes.GetBody()); err != nil {
				resultCh <- []m2cp.RpcResultReadonly{
					rpctypes.NewRpcResultError(fmt.Sprintf("failed parsing CoAP response: %s", err), m2cp.RpcErrorCodeFailed),
				}
				return
			}
			resMsg.AddTrace(n.GetAddress())
			resultCh <- resMsg.GetResults()
		}()
	} else {
		// Send via AMQP

		// Channel for receiving response from transport layer
		responseCh := make(chan []m2cp.RpcResultReadonly, 1)
		n.pendingRPCsLock.Lock()
		n.pendingRPCs[cmdMsgId] = responseCh
		n.pendingRPCsLock.Unlock()

		n.ctp.LogDebug("[RPC %s] sending to %s", cmdMsgId, to.GetAddress())
		err = n.networkConnection.SendMessage(cmdMsg)
		if err != nil {
			return nil, err
		}

		// Start a Goroutine to wait for a response or time out
		go func() {
			timeout := time.NewTimer(n.rpcTimeout) // Set your desired timeout duration
			defer timeout.Stop()

			select {
			case result := <-responseCh:
				//n.ctx.LogInfo("[RPC %s] got response from channel", cmdMsgId)
				resultCh <- result
			case <-timeout.C:
				n.ctp.LogWarn("[RPC %s] got timeout", cmdMsgId)
				n.pendingRPCsLock.Lock()
				delete(n.pendingRPCs, cmdMsgId)
				n.pendingRPCsLock.Unlock()
				resultCh <- []m2cp.RpcResultReadonly{
					rpctypes.NewRpcResultError("RPC timed out", m2cp.RpcErrorCodeRequestTimeout),
				}
			case <-n.ctp.Done():
				//n.ctx.LogInfo("[RPC %s] got context done", cmdMsgId)
				break
			}
			n.pendingRPCsLock.Lock()
			delete(n.pendingRPCs, cmdMsgId)
			n.pendingRPCsLock.Unlock()
			//n.ctx.LogInfo("[RPC %s] closing response channel", cmdMsgId)
			close(responseCh)
			close(resultCh)
			//n.ctx.LogInfo("[RPC %s] done with this rpc", cmdMsgId)

		}()
	}

	// Return the response channel
	return resultCh, nil
}

func (n *node) SetRpcTimeout(timeout time.Duration) {
	const minTimeout = 1 * time.Second
	if timeout < minTimeout {
		n.rpcTimeout = minTimeout
	} else {
		n.rpcTimeout = timeout
	}
}

func (n *node) handleRpcCall(cmd m2cp.CommandMessage) (m2cp.ResponseMessage, error) {
	if n.commandHandler == nil {
		return nil, fmt.Errorf("no RPC handler set")
	}
	return n.commandHandler.Process(n.ctp, cmd)
}

func (n *node) handleRpcResponse(res m2cp.ResponseMessage) {
	n.pendingRPCsLock.Lock()
	responseCh, ok := n.pendingRPCs[res.GetCommandId()]
	if ok {
		responseCh <- res.GetResults()
	} else {
		n.ctp.LogWarn("rpc response not found for id %s", res.GetHeader().GetId())
	}
	n.pendingRPCsLock.Unlock()
}

func (n *node) calcDataQueueMaxAge(format m2cp.DataFormat) time.Duration {
	if format.GetDataQueueMaxAge() < n.dataQueueMaxAge {
		return format.GetDataQueueMaxAge()
	}
	return n.dataQueueMaxAge
}

func (n *node) calcDataQueueMaxCount(format m2cp.DataFormat) int {
	if format.GetDataQueueMaxCount() < n.dataQueueMaxCount {
		return format.GetDataQueueMaxCount()
	}
	return n.dataQueueMaxCount
}

func (n *node) dataQueueManage(format m2cp.DataFormat, queue *dataQueue) {
	timer := time.NewTimer(n.calcDataQueueMaxAge(format))
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
				// We successfully drained the channel
			default:
				// The channel was already empty
			}
		}
	}()

	for {
		if len(queue.rows) == 0 {
			// if queue is empty, we can wait for new data without having a timer set
			select {
			case data := <-queue.ch:
				//n.ctx.LogDebug("got data row for empty queue")
				n.dataQueueAppend(queue, data)
				if queue.size >= n.calcDataQueueMaxCount(format) {
					//n.ctx.LogDebug("data queue full -> sending now, size: %d", queue.size)
					n.dataQueueSend(format, queue)
				}
			case <-n.ctp.Done():
				return
			}
		} else {
			// wait for new data or timeout
			select {
			case data := <-queue.ch:
				//n.ctx.LogDebug("got data row for non-empty queue")
				n.dataQueueAppend(queue, data)
				if queue.size >= n.calcDataQueueMaxCount(format) {
					//n.ctx.LogDebug("data queue full -> sending now, size: %d", queue.size)
					n.dataQueueSend(format, queue)
				}
			case <-timer.C:
				//n.ctx.LogDebug("data queue timeout, size: %d", queue.size)
				if len(queue.rows) > 0 {
					n.dataQueueSend(format, queue)
				}
			case <-n.ctp.Done():
				return
			}
		}

		// Reset the timer after processing data
		if !timer.Stop() {
			select {
			case <-timer.C:
				// We successfully drained the channel
			default:
				// The channel was already empty, which we didn't expect
			}
		}
		timer.Reset(n.calcDataQueueMaxAge(format))
	}
}

func (n *node) dataQueueAppend(queue *dataQueue, data m2cp.DataRow) {
	queue.rows = append(queue.rows, data)
	queue.size++
	queue.lastUpdated = time.Now()
	//n.ctx.LogDebug("added data to queue, size: %d", queue.size)
}

func (n *node) dataQueueSend(format m2cp.DataFormat, stack *dataQueue) {
	// n.ctx.LogDebug("sending data queue, size: %d", stack.size)
	n.dataQueuesLock.Lock()
	rows := stack.rows
	stack.rows = []m2cp.DataRow{}
	stack.size = 0
	n.dataQueuesLock.Unlock()

	msg, err := messages.NewDataMessage(n, format)
	if err != nil {
		n.ctp.LogError("Failed to create data message: %s", err.Error())
		return
	}
	msg.AddRows(rows)
	err = n.networkConnection.SendMessage(msg)
	if err != nil {
		FailedToSendCount++
		n.ctp.LogError("Failed to send data message: %s", err.Error())
	}
}

var FailedToSendCount = 0

func (n *node) GetFailedToSendCount() int {
	return FailedToSendCount
}
