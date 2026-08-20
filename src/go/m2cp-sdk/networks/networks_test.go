package networks

import (
	"fmt"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/messages"
	"m2cp/rpc"
	"m2cp/rpc/rpctypes"
	"m2cp/tools"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func (s *TestSuite) TestSendSubscribeReceiveSignals() {
	con, err := NewNetworkConnection(s.ctp)
	defer func() {
		if con != nil {
			con.Close()
		}
	}()

	s.NotNil(con)
	s.Nil(err)
	s.ctp.LogDebug(fmt.Sprintf("domain: %s", con.GetDomain().GetName()))

	con.AwaitReady()
	s.ctp.LogInfo("---------------------- CONNECTED ----------------------")

	testId := "test-" + uuid.New().String()

	received := 0
	handleSignal := func(signal m2cp.SignalMessage) {
		if signal.GetName() != testId {
			return
		}
		received++
	}
	err = con.SubscribeSignals([]string{"*"}, handleSignal)
	if err != nil {
		s.Nil(err)
	}
	con.AwaitReady()
	s.ctp.LogInfo("---------------------- SUBSCRIBED ----------------------")

	nodeA, err := con.NewNode("test")
	s.NotNil(nodeA)
	s.Nil(err)
	s.ctp.LogInfo("node address: %s", nodeA.GetAddress())

	const numSignals = 500
	timeStart := time.Now()

	sent := 0
	for i := 0; i < numSignals; i++ {
		err = nodeA.EmitSignal(testId, fmt.Sprintf("sent by Go-SDK at %s", time.Now().String()), messages.SignalType_Info)
		sent++
		s.Nil(err)
	}

	for received < sent {
		s.ctp.Sleep(1 * time.Second)
		if time.Now().Sub(timeStart) > 10*time.Second {
			s.ctp.Cancel()
			s.Fail("failed to receive all signals, expected %d, received %d", sent, received)
			return
		}
	}

	s.ctp.LogInfo("received %d signals in %s, %d msg/sec", received, time.Now().Sub(timeStart).String(), int(float64(received)/time.Now().Sub(timeStart).Seconds()))
}

type sendSubscribeReceiveOptions struct {
	NumSignals               int
	StopAndRestartMessageHub bool
	DelayBetweenSignals      time.Duration
	ReportMemoryUsage        bool
	MessageSerializer        m2cp.MessageSerializer
}

// TestSendSubscribeReceiveSignalsPersistent tests sending signals to a subscriber who is offline and then comes online
func (s *TestSuite) TestSendSubscribeReceiveSignalsMini() {
	s.sendSubscribeReceiveSignals(sendSubscribeReceiveOptions{
		NumSignals:        3,
		MessageSerializer: m2cp.MessageSerializerJson,
	})
}

// TestSendSubscribeReceiveSignalsPersistent tests sending signals to a subscriber who is offline and then comes online
func (s *TestSuite) TestSendSubscribeReceiveSignalsSerializerNone() {
	s.sendSubscribeReceiveSignals(sendSubscribeReceiveOptions{
		NumSignals:        3000,
		MessageSerializer: m2cp.MessageSerializerJson,
	})
}

// TestSendSubscribeReceiveSignalsPersistent tests sending signals to a subscriber who is offline and then comes online
func (s *TestSuite) TestSendSubscribeReceiveSignalsPersistent() {
	s.sendSubscribeReceiveSignals(sendSubscribeReceiveOptions{
		NumSignals:        1000,
		MessageSerializer: m2cp.MessageSerializerDefault,
	})
}

// TestSendSubscribeReceiveSignalsPersistent tests sending signals to a subscriber who is offline and then comes online
/*func (s *TestSuite) TestSendSubscribeReceiveSignalsPersistentLoadTest() {
	s.ctp.SetLogLevel(m2cp.LogLevelWarn)
	s.sendSubscribeReceiveSignals(sendSubscribeReceiveOptions{
		NumSignals:          10000,
		DelayBetweenSignals: 2 * time.Millisecond,
		ReportMemoryUsage:   true,
	})
}
*/

func (s *TestSuite) TestSubscribeReceiveSignalPersistentRestartHubOne() {
	s.sendSubscribeReceiveSignals(sendSubscribeReceiveOptions{
		NumSignals:               100,
		StopAndRestartMessageHub: true,
		MessageSerializer:        m2cp.MessageSerializerDefault,
	})
}

func (s *TestSuite) TestSubscribeReceiveSignalPersistentRestartHub() {
	s.sendSubscribeReceiveSignals(sendSubscribeReceiveOptions{
		NumSignals:               5000,
		StopAndRestartMessageHub: true,
		MessageSerializer:        m2cp.MessageSerializerDefault,
	})
}

func (s *TestSuite) sendSubscribeReceiveSignals(options sendSubscribeReceiveOptions) {
	ctpMem := s.ctp.BranchWithName("memory")
	ctpMem.SetLogLevel(m2cp.LogLevelInfo)

	var (
		sent          = 0
		received      = 0
		conSubscriber m2cp.NetworkConnection
		err           error
		timeStart     time.Time
	)
	const (
		subscriptionId = "persistent-signal-test-subscription"
		signalTopic    = "persistent-signal-test"
	)

	handleSignal := func(signal m2cp.SignalMessage) {
		if signal.GetName() == signalTopic {
			s.ctp.LogInfo("received signal topic=%s", signal.GetHeader().GetTopic())
			received++
		}
	}

	senderNodeName := fmt.Sprintf("test-%s", uuid.New().String())

	startSubscriber := func() {
		s.T().Logf("----------------------[ startSubscriber ]----------------------")
		conSubscriber, err = NewNetworkConnectionWithOptions(s.ctp, m2cp.NetworkConnectionOptions{
			AppName: "persistent-subscriber",
		})
		s.NotNil(conSubscriber)
		s.Nil(err)
		s.ctp.LogInfo(fmt.Sprintf("domain: %s", conSubscriber.GetDomain().GetName()))

		err = conSubscriber.SubscribeSignalsWithOptions([]string{senderNodeName}, handleSignal, m2cp.SubscriptionOptions{
			PersistenceId: tools.StrPtr(subscriptionId),
		})
		if err != nil {
			s.Nil(err)
		}

		conSubscriber.AwaitReady()
	}
	stopSubscriber := func() {
		s.T().Logf("----------------------[ stopSubscriber ]----------------------")
		conSubscriber.Close()
	}

	var conSender m2cp.NetworkConnection
	startRunStopSender := func() {
		var err error
		s.T().Logf("----------------------[ startRunStopSender ]----------------------")
		conSender, err = NewNetworkConnectionWithOptions(s.ctp, m2cp.NetworkConnectionOptions{
			MessageSerializer: options.MessageSerializer,
		})
		s.NoError(err)
		s.NotNil(conSender)
		nodeA, err := conSender.NewNode(senderNodeName)
		s.NotNil(nodeA)
		s.Nil(err)
		s.ctp.LogInfo("node address: %s", nodeA.GetAddress())

		conSender.AwaitReady()

		timeStart = time.Now()

		sent = 0
		for i := 0; i < options.NumSignals; i++ {
			s.ctp.LogInfo("sending signal #%d from node #%d '%s'", i, 0, nodeA.GetName())
			for retryEmit := 0; retryEmit < 3; retryEmit++ {
				if err = nodeA.EmitSignal(signalTopic, time.Now().String(), m2cp.SignalLevelInfo); err != nil {
					s.ctp.Sleep(10 * time.Millisecond)
					continue
				}
				break
			}
			if options.DelayBetweenSignals > 0 {
				s.ctp.Sleep(options.DelayBetweenSignals)
			}
			sent++
			if sent%100 == 0 {
				ctpMem.LogInfo("sent: %d, mem: %s", sent, tools.GetMemoryUsageRssHumanized())
			}
			s.Nil(err)
		}
		//s.ctp.Sleep(5 * time.Second)
		conSender.Close()
	}

	countIncomingMessages := func() {
		s.T().Logf("----------------------[ countIncomingMessages ]----------------------")
		waitingSince := time.Now()
		for received < sent && (time.Now().Sub(waitingSince) < 5*time.Second) {
			s.ctp.LogInfo("waiting %d seconds, received %d/%d signals", int(time.Now().Sub(waitingSince).Seconds()), received, sent)
			s.ctp.Sleep(1 * time.Second)
			if time.Now().Sub(timeStart) > 30*time.Second {
				s.Fail("failed to receive all signals")
				return
			}
		}
		s.ctp.LogInfo("received %d/%d signals", received, sent)
	}

	startSubscriber()
	stopSubscriber()
	if options.StopAndRestartMessageHub {
		s.restartMessageHub()
	}

	ctpMem.LogInfo(tools.GetMemoryUsageRssHumanized())

	startRunStopSender()
	startSubscriber()
	countIncomingMessages()
	stopSubscriber()

	ctpMem.LogInfo(tools.GetMemoryUsageRssHumanized())

	senderStats := conSender.GetStats().GetSenderStats()
	s.GreaterOrEqual(int(senderStats.GetTransmitted()), sent)
	s.GreaterOrEqual(int(senderStats.GetTransmitted()), sent)
	s.Zero(senderStats.GetErrors())
	s.Zero(senderStats.GetQueueSendLen())
	s.Zero(senderStats.GetQueueAckLen())

	s.ctp.LogInfo("received %d signals in %s, %d msg/sec", received, time.Now().Sub(timeStart).String(), int(float64(received)/time.Now().Sub(timeStart).Seconds()))
	s.Equal(sent, received, "didn't receive all signals")
}

func (s *TestSuite) TestSubscribeSpecificTopics() {
	tests := []struct {
		topicSubscribe string
		topicsEmit     []string
		shouldMatch    int
	}{
		{"*", []string{"a", "b", "c"}, 3},
		{"a", []string{"a", "b", "c"}, 1},
		{"a.*", []string{"a.x", "a.y", "a"}, 2},

		// NOTE: LavinMQ handles "a.*" and "a.#" the same way, so the following tests will fail
		// NOTE: RabbitMQ handles "a.*" and "a.#" differently, so the following tests will pass
		//{"a.*", []string{"ax", "a.x", "a.x.foo", "a.x.bar"}, 1},

		{"a.*.00000000-0000-0000-0000-000000000000", []string{"ax", "a.x", "a.x.00000000-0000-0000-0000-000000000000", "a.x.11111111-1111-1111-1111-111111111111"}, 1},
		{"a.#", []string{"ax", "a.x", "a.x.00000000-0000-0000-0000-000000000000", "a.x.11111111-1111-1111-1111-111111111111"}, 3},
		{"a.*.*", []string{"ax", "a.x", "a.x.00000000-0000-0000-0000-000000000000", "a.x.11111111-1111-1111-1111-111111111111"}, 2},
	}

	con, err := NewNetworkConnection(s.ctp)
	defer con.Close()
	s.NotNil(con)
	s.Nil(err)
	con.AwaitReady()

	conEmit, err := NewNetworkConnection(s.ctp)
	defer conEmit.Close()
	s.NotNil(conEmit)
	s.Nil(err)
	conEmit.AwaitReady()

	s.ctp.LogInfo(fmt.Sprintf("domain con: %s", con.GetDomain().GetName()))
	s.ctp.LogInfo(fmt.Sprintf("domain conEmit: %s", conEmit.GetDomain().GetName()))

	for i, test := range tests {
		s.ctp.LogInfo("\n\n------------------- TEST iteration %d", i)

		// to avoid counting stray signals, we create a unique id to identify signals from this test iteration
		testId := "test-" + uuid.New().String()

		received := 0
		handlerId := i
		handleSignal := func(signal m2cp.SignalMessage) {
			if signal.GetName() != testId {
				return
			}
			s.ctp.LogInfo("handler %d received signal topic=%s", handlerId, signal.GetHeader().GetTopic())
			received++
		}

		testContext := contextplus.NewContextPlus()
		err = con.SubscribeSignalsWithOptions(
			[]string{test.topicSubscribe}, handleSignal, m2cp.SubscriptionOptions{Context: testContext})
		if err != nil {
			s.Nil(err)
		}
		con.AwaitReady()

		nodes := make([]m2cp.Node, len(test.topicsEmit))
		for i, topic := range test.topicsEmit {
			s.ctp.LogInfo("creating node #%d:  %s", i, topic)
			nodes[i], err = conEmit.NewNode(topic)
			s.ctp.LogInfo("node address: %s", nodes[i].GetAddress())
			s.NotNil(nodes[i])
			s.Nil(err)
			conEmit.AwaitReady()
		}

		const numSignalsPerNode = 3
		timeStart := time.Now()

		sent := 0
		for i := 0; i < numSignalsPerNode; i++ {
			for j, node := range nodes {
				s.ctp.LogInfo("sending signal #%d from node #%d '%s'", i, j, node.GetName())
				err = node.EmitSignal(testId, fmt.Sprintf("sent by Go-SDK at %s", time.Now().String()), messages.SignalType_Info)
				sent++
				s.Nil(err)
			}
		}

		for received < test.shouldMatch*numSignalsPerNode {
			s.ctp.Sleep(1 * time.Second)
			if time.Now().Sub(timeStart) > 30*time.Second {
				s.ctp.Cancel()
				s.Fail("failed to receive all signals")
				return
			}
		}
		s.ctp.Sleep(1 * time.Second)

		if received > test.shouldMatch*numSignalsPerNode {
			s.Fail("received too many signals")
		}

		s.ctp.LogInfo("received %d signals in %s, %d msg/sec", received, time.Now().Sub(timeStart).String(), int(float64(received)/time.Now().Sub(timeStart).Seconds()))

		// cancel subscriptions
		testContext.Cancel()
	}
}

func (s *TestSuite) TestSubscriberAutoRestore() {
	var (
		numSignals           = 3
		sent                 = 0
		received             = 0
		durationWaitIncoming = 10 * time.Second
		conSubscriber        m2cp.NetworkConnection
		err                  error
		timeStart            time.Time
		randomTopic          = "test-" + uuid.New().String()
	)

	handleSignal := func(signal m2cp.SignalMessage) {
		if signal.GetName() == randomTopic {
			s.ctp.LogInfo("received signal topic=%s", signal.GetHeader().GetTopic())
			received++
		}
	}

	startSubscriber := func() {
		s.T().Logf("----------------------[ startSubscriber ]----------------------")
		conSubscriber, err = NewNetworkConnectionWithOptions(s.ctp, m2cp.NetworkConnectionOptions{
			AppName: "subscriber",
		})
		s.NotNil(conSubscriber)
		s.Nil(err)
		s.ctp.LogInfo(fmt.Sprintf("domain: %s", conSubscriber.GetDomain().GetName()))

		err = conSubscriber.SubscribeSignals([]string{"*"}, handleSignal)
		if err != nil {
			s.Nil(err)
		}

		conSubscriber.AwaitReady()
	}
	stopSubscriber := func() {
		s.T().Logf("----------------------[ stopSubscriber ]----------------------")
		conSubscriber.Close()
	}

	startSender := func() {
		s.T().Logf("----------------------[ startSender ]----------------------")
		conSender, err := NewNetworkConnection(s.ctp)
		s.NoError(err)
		s.NotNil(conSender)
		nodeA, err := conSender.NewNode("test")
		s.NotNil(nodeA)
		s.Nil(err)
		s.ctp.LogInfo("node address: %s", nodeA.GetAddress())

		conSender.AwaitReady()

		timeStart = time.Now()

		sent = 0
		for i := 0; i < numSignals; i++ {
			select {
			case <-s.ctp.Done():
				break
			default:
				s.ctp.LogInfo("sending signal #%d from node #%d '%s'", sent, 0, nodeA.GetName())
				err = nodeA.EmitSignal(randomTopic, time.Now().String(), m2cp.SignalLevelInfo)
				s.Nil(err)
				sent++
				s.ctp.Sleep(1000 * time.Millisecond)
			}
		}
		s.T().Logf("----------------------[ stopSender ]----------------------")
		conSender.Close()
	}

	countIncomingMessages := func() {
		s.T().Logf("----------------------[ countIncomingMessages ]----------------------")
		waitingSince := time.Now()
		for received < sent && (time.Now().Sub(waitingSince) < 5*time.Second) {
			s.ctp.LogInfo("waiting %d seconds, received %d/%d signals", int(time.Now().Sub(waitingSince).Seconds()), received, sent)
			s.ctp.Sleep(1 * time.Second)
			if time.Now().Sub(timeStart) > durationWaitIncoming {
				s.Fail("failed to receive all signals")
				return
			}
		}
		s.ctp.LogInfo("received %d/%d signals", received, sent)
	}

	startSubscriber()
	s.restartMessageHub()
	startSender()
	countIncomingMessages()
	stopSubscriber()

	s.ctp.LogInfo("received %d signals in %s, %d msg/sec", received, time.Now().Sub(timeStart).String(), int(float64(received)/time.Now().Sub(timeStart).Seconds()))
	s.Equal(sent, received, "didn't receive all signals")
}

/**
 * Test subscribing signals from a specific node
 */

//func (s *TestSuite) TestSendSubscribeReceiveSignalsTopic() {
//	con, err := NewNetworkConnection(s.ctx)
//	defer con.Close()
//	s.NotNil(con)
//	s.Nil(err)
//	s.ctx.LogInfo(fmt.Sprintf("domain: %s", con.GetDomain().GetName()))
//
//	testId := "test-" + uuid.New().String()
//
//	nodeA, err := con.NewNode("test")
//	s.NotNil(nodeA)
//	s.Nil(err)
//	s.ctx.LogInfo("node address: %s", nodeA.GetAddress())
//
//	received := 0
//	handleSignal := func(signal m2cp.SignalMessage) {
//		if signal.GetName() != testId {
//			return
//		}
//		received++
//	}
//	err = con.SubscribeSignals([]string{nodeA.GetAddress()}, handleSignal)
//	if err != nil {
//		s.Nil(err)
//	}
//
//	const numSignals = 3
//	timeStart := time.Now()
//
//	sent := 0
//	for i := 0; i < numSignals; i++ {
//		err = nodeA.EmitSignal(testId, fmt.Sprintf("sent by Go-SDK at %s", time.Now().String()), messages.SignalType_Info)
//		sent++
//		s.Nil(err)
//	}
//
//	for received < sent {
//		s.ctx.Sleep(1 * time.Second)
//		if time.Now().Sub(timeStart) > 5*time.Second {
//			s.ctx.Cancel()
//			s.Fail("failed to receive all signals")
//			return
//		}
//	}
//
//	s.ctx.LogInfo("received %d signals in %s, %d msg/sec", received, time.Now().Sub(timeStart).String(), int(float64(received)/time.Now().Sub(timeStart).Seconds()))
//}

func (s *TestSuite) TestSendSubscribeReceiveData() {

	type config struct {
		MaxCount int
		MaxAge   time.Duration
		Rows     int
	}
	configs := []config{
		{MaxCount: 1000, MaxAge: 10 * time.Millisecond, Rows: 1000},
		//{MaxCount: 1, MaxAge: 1 * time.Second, Rows: 1},
		//{MaxCount: 1, MaxAge: 1 * time.Second, Rows: 3},
		//{MaxCount: 1, MaxAge: 1 * time.Second, Rows: 1000},
		//{MaxCount: 10, MaxAge: 1 * time.Second, Rows: 1000},
		//{MaxCount: 100, MaxAge: 1 * time.Second, Rows: 1000},
		//{MaxCount: 100, MaxAge: 10 * time.Second, Rows: 10},
		//{MaxCount: 1000, MaxAge: 500 * time.Millisecond, Rows: 5000},
	}
	for i, c := range configs {
		s.ctp.LogInfo("round %d/%d: maxCount=%d, maxAge=%s, rows=%d", i, len(configs), c.MaxCount, c.MaxAge.String(), c.Rows)
		s.testSendSubscribeReceiveData(c.MaxCount, c.MaxAge, c.Rows)
		s.ctp.LogInfo("round %d passed", i)
		s.ctp.Sleep(1 * time.Second)
	}
}

func (s *TestSuite) TestNetworkDiscovery() {
	s.ctp.SetLogLevel(m2cp.LogLevelInfo)

	con0, err := NewNetworkConnection(s.ctp)
	defer con0.Close()

	con1, err := NewNetworkConnection(s.ctp)
	defer con1.Close()

	nodeA1, err := con0.NewNode("a1")
	s.NotNil(nodeA1)
	s.Nil(err)

	con2, err := NewNetworkConnection(s.ctp)
	defer con2.Close()
	s.NoError(err)
	s.NotNil(con2)

	con2.GetKnownNodeAddresses()

	s.NoError(con2.DiscoverNodes()) // NOTE: it doesn't matter, which node does the discovery - all nodes will know all nodes

	s.ctp.Sleep(1 * time.Second)

	// all connections should know all nodes now
	type check struct {
		con     m2cp.NetworkConnection
		success bool
	}

	checks := []check{
		{con: con0, success: false},
		{con: con1, success: false},
		{con: con2, success: false},
	}
	const maxTries = 10
	for tries := 0; tries < maxTries; tries++ {
		s.ctp.LogInfo("try %d/%d", tries, maxTries)
		allGood := true
		for i := range checks {
			c := checks[i]
			adds := c.con.GetKnownNodeAddresses()
			for _, add := range adds {
				if add == nodeA1.GetAddress() {
					c.success = true
				}
			}
			s.ctp.LogInfo("connection %d discovered A1: %v", i, c.success)
			if !c.success {
				allGood = false
				s.NoError(c.con.DiscoverNodes())
				s.ctp.Sleep(1 * time.Second)
			}
		}
		if allGood {
			s.ctp.LogInfo("all connections know all nodes")
			return
		}
	}
	s.Fail("not all connections know all nodes")
}

func (s *TestSuite) testSendSubscribeReceiveData(maxCount int, maxAge time.Duration, numRows int) {
	s.ctp.SetLogLevel(m2cp.LogLevelInfo)
	con, err := NewNetworkConnection(s.ctp)
	defer func() {
		con.Close()
		s.ctp.Sleep(3 * time.Second)
	}()

	s.NotNil(con)
	s.Nil(err)
	s.ctp.LogInfo(fmt.Sprintf("domain: %s", con.GetDomain().GetName()))

	testId := "test-" + uuid.New().String()
	received := 0

	handleData := func(data m2cp.DataMessage) {
		if data.GetFormat().GetFormatId() != testId {
			return
		}
		received = received + data.CountRows()

		for _, row := range data.GetRows() {
			s.Nil(row.GetFieldInt("nillable"))
		}
	}
	err = con.SubscribeData([]string{"*"}, handleData)
	if err != nil {
		s.Nil(err)
	}

	nodeSender, err := con.NewNode("test")
	s.NotNil(nodeSender)
	s.Nil(err)
	s.ctp.LogInfo("node address: %s", nodeSender.GetAddress())
	nodeSender.SetDataQueueMaxCount(maxCount)
	nodeSender.SetDataQueueMaxAge(maxAge)

	timeStart := time.Now()

	format, err := messages.NewDataFormat(testId,
		messages.NewDataFieldBool("foo"),
		messages.NewDataFieldInt("counter"),
		messages.NewDataFieldInt("nillable"),
	)
	s.NoError(err)

	con.AwaitReady()

	sent := 0
	for i := 0; i < numRows; i++ {
		//err = nodeA.EmitSignal(testId, fmt.Sprintf("sent by Go-SDK at %s", time.Now().String()), messages.SignalType_Info)
		var row m2cp.DataRow
		row, err = format.NewRow(map[string]interface{}{
			"foo":      true,
			"counter":  i,
			"nillable": (*int)(nil),
		})
		s.NoError(err)
		err = nodeSender.EmitDataRow(row)
		if err != nil {
			s.Fail(err.Error())
			return
		}
		sent++
		s.ctp.Sleep(1 * time.Millisecond)

	}

	for received < sent {
		if time.Now().Sub(timeStart) > 60*time.Second {
			s.ctp.Cancel()
			s.Fail(fmt.Sprintf("failed to receive all data (sent: %d, received: %d)", sent, received))
			return
		}
		s.ctp.Sleep(1 * time.Second)
	}
	s.ctp.LogInfo("received %d data rows in %s, %d rows/sec", received, time.Now().Sub(timeStart).String(), int(float64(received)/time.Now().Sub(timeStart).Seconds()))
}

func (s *TestSuite) TestSnapd() {
	s.ctp.SetLogLevel(m2cp.LogLevelInfo)
	con, err := NewNetworkConnection(s.ctp)
	defer con.Close()
	s.NotNil(con)
	s.Nil(err)
	s.ctp.LogInfo(fmt.Sprintf("domain: %s", con.GetDomain().GetName()))
}

func (s *TestSuite) TestRpcTimout() {
	con, err := NewNetworkConnection(s.ctp)
	defer con.Close()
	s.NotNil(con)
	s.Nil(err)

	n, err := con.NewNode("test-node")
	s.NotNil(n)
	s.Nil(err)
	n.SetRpcTimeout(1 * time.Second)

	addr, err := messages.NewAddress(s.ctp, "no-node.no-app.local")
	s.NoError(err)
	resChan, err := n.RemoteProcedureCall(addr, "some-command", nil)
	s.NotNil(resChan)
	s.Nil(err)

	for res := range resChan {
		s.False(res.IsSuccessful())
		s.Contains(res.GetMessage(), "timed out")
		s.Equal(m2cp.RpcErrorCodeRequestTimeout, res.GetErrorCode())
	}
}

func (s *TestSuite) TestRpcNodeUptime() {

	conA, err := NewNetworkConnection(s.ctp)
	s.NotNil(conA)
	s.Nil(err)
	defer conA.Close()

	nodeA, err := conA.NewNode("rpc-endpoint")
	s.NotNil(nodeA)
	s.Nil(err)
	handler, err := rpc.NewRpcHandler()
	s.Nil(err)
	nodeA.SetRpcHandler(handler)

	conA.AwaitReady()

	conB, err := NewNetworkConnection(s.ctp)
	defer conB.Close()
	s.NotNil(conB)
	s.Nil(err)

	nodeB, err := conB.NewNode("rpc-caller")
	s.NotNil(nodeB)
	s.Nil(err)
	nodeB.SetRpcTimeout(15 * time.Second)

	conB.AwaitReady()

	resChan, err := nodeB.RemoteProcedureCall(nodeA, "NodeUptime", map[string]string{})

	s.Nil(err)
	s.NotNil(resChan)
	receivedResults := 0
	for res := range resChan {
		receivedResults++
		s.NotEmpty(res.GetString("sdk", ""))
		s.NotEqual(-1, res.GetInt("uptime", -1))
	}
	s.Equal(1, receivedResults)
}

func (s *TestSuite) TestRpcOnce() {
	s.testRpcOnce(1)
}

func (s *TestSuite) TestRpc() {
	const repetitions = 10
	for i := 0; i < repetitions; i++ {
		fmt.Printf("\n\n--[Iteration %d/%d]--------------------------------------------------------------------------\n", i, repetitions)
		s.testRpcOnce(2)
	}
}

// rmq.DebugSender = true
// rmq.DebugSubscriber = true
func (s *TestSuite) testRpcOnce(numCalls int) {

	conA, err := NewNetworkConnection(s.ctp)
	s.NotNil(conA)
	s.Nil(err)
	defer conA.Close()

	nodeA, err := conA.NewNode("rpc-endpoint")
	s.NotNil(nodeA)
	s.Nil(err)
	handler, err := rpc.NewRpcHandler(exampleCommand())
	s.Nil(err)
	nodeA.SetRpcHandler(handler)

	conA.AwaitReady()

	conB, err := NewNetworkConnection(s.ctp)
	defer conB.Close()
	s.NotNil(conB)
	s.Nil(err)

	nodeB, err := conB.NewNode("rpc-caller")
	s.NotNil(nodeB)
	s.Nil(err)
	nodeB.SetRpcTimeout(15 * time.Second)

	conB.AwaitReady()

	for i := 0; i < numCalls; i++ {
		resChan, err := nodeB.RemoteProcedureCall(nodeA, "Test", map[string]string{
			"x":  "42",
			"xs": "1,2,3,4,5",
		})

		s.Nil(err)
		s.NotNil(resChan)
		receivedResults := 0
		for res := range resChan {
			receivedResults++
			s.Equal("yeah", res.GetMessage())
			s.Equal(3.141592653589793, res.GetDouble("d", 0))
			s.Equal(42, res.GetInt("y", 0))
			s.Len(res.GetTraces(), 5)
		}
		s.Equal(1, receivedResults)
	}
}

func (s *TestSuite) TestRpcLocal() {
	ctx := s.ctp
	assert.NoError(s.T(), func() error {
		conA, err := NewNetworkConnection(s.ctp)
		s.NotNil(conA)
		s.Nil(err)
		defer conA.Close()

		nodeA, err := conA.NewNode("rpc-endpoint")
		s.NotNil(nodeA)
		s.Nil(err)
		handler, err := rpc.NewRpcHandler(exampleCommand())
		s.Nil(err)
		nodeA.SetRpcHandler(handler)

		conB, err := NewNetworkConnection(ctx)
		if err != nil {
			return err
		}

		nodeB, err := conB.NewNode("any-node-name-for-rpc-caller")
		if err != nil {
			return err
		}

		conA.AwaitReady()
		conB.AwaitReady()

		// Assuming, that the endpoint we're calling is called "endpoint" and is part of the app "appname"
		recipient, err := messages.NewAddress(s.ctp, fmt.Sprintf("%s.%s.local", nodeA.GetNodeName(), nodeA.GetAppName()))
		s.NoError(err)
		resChan, err := nodeB.RemoteProcedureCall(recipient, "Test", map[string]string{
			"x":  "42",
			"xs": "1,2,3,4,5",
		})
		s.Nil(err)
		s.NotNil(resChan)
		receivedResults := 0
		for res := range resChan {
			receivedResults++
			s.Equal("yeah", res.GetMessage())
			s.Equal(3.141592653589793, res.GetDouble("d", 0))
			s.Equal(42, res.GetInt("y", 0))
			s.Len(res.GetTraces(), 5)
		}
		s.Equal(1, receivedResults)

		return nil
	}())
}

// TestRpcVirtualDevice sends a rpc call to m2cp-gateway in the virtual device (must be configured with M2CP_VIRTUAL_DEVICE)
func (s *TestSuite) TestRpcVirtualDevice() {
	ctx := s.ctp
	assert.NoError(s.T(), func() error {
		conB, err := NewNetworkConnection(ctx)
		if err != nil {
			return err
		}

		nodeB, err := conB.NewNode("foobar")
		if err != nil {
			return err
		}

		conB.AwaitReady()

		// Assuming, that the endpoint we're calling is called "endpoint" and is part of the app "appname"
		recipient, err := messages.NewAddress(s.ctp, "rpc.m2cp-gateway.local")
		s.NoError(err)
		resChan, err := nodeB.RemoteProcedureCall(recipient, "NodeUptime", nil)
		s.Nil(err)
		s.NotNil(resChan)
		receivedResults := 0
		for res := range resChan {
			receivedResults++
			s.True(res.IsSuccessful())
			s.ctp.LogInfo("rpc result message: " + res.GetMessage())
		}
		s.Equal(1, receivedResults)

		return nil
	}())
}

func (s *TestSuite) TestRpcLoooongResponse() {
	conA, err := NewNetworkConnection(s.ctp)
	s.NotNil(conA)
	s.Nil(err)
	defer conA.Close()

	nodeA, err := conA.NewNode("rpc-endpoint")
	s.NotNil(nodeA)
	s.Nil(err)
	handler, err := rpc.NewRpcHandler(exampleCommandLoooongResponse())
	s.Nil(err)
	nodeA.SetRpcHandler(handler)

	//wait 3 seconds
	s.ctp.Sleep(3 * time.Second)

	conB, err := NewNetworkConnection(s.ctp)
	defer conB.Close()
	s.NotNil(conB)
	s.Nil(err)

	nodeB, err := conB.NewNode("rpc-caller")
	s.NotNil(nodeB)
	s.Nil(err)
	nodeB.SetRpcTimeout(5 * time.Second)

	for i := 0; i < 10; i++ {
		maxMessageSize := i * 10000
		s.ctp.LogInfo("maxMessageSize: %d", maxMessageSize)
		resChan, err := nodeB.RemoteProcedureCall(nodeA, "Logs", map[string]string{
			"len": fmt.Sprintf("%d", maxMessageSize),
		})

		s.Nil(err)
		s.NotNil(resChan)
		receivedResults := 0
		for res := range resChan {
			receivedResults++
			logsLen := len(res.GetString("logs", ""))
			if maxMessageSize > logsLen {
				s.Fail("response too short")
			}
		}
		s.Equal(1, receivedResults)
	}
}

func (s *TestSuite) TestRpcResponseRewrite() {
	conA, err := NewNetworkConnection(s.ctp)
	s.NotNil(conA)
	s.Nil(err)
	defer conA.Close()

	nodeA, err := conA.NewNode("rpc-endpoint")
	s.NotNil(nodeA)
	s.Nil(err)
	handler, err := rpc.NewRpcHandler(exampleCommand())
	s.Nil(err)
	nodeA.SetRpcHandler(handler)

	//wait 3 seconds
	s.ctp.Sleep(3 * time.Second)

	conB, err := NewNetworkConnection(s.ctp)
	defer conB.Close()
	s.NotNil(conB)
	s.Nil(err)

	nodeB, err := conB.NewNode("rpc-caller")
	s.NotNil(nodeB)
	s.Nil(err)
	nodeB.SetRpcTimeout(5 * time.Second)

	resChan, err := nodeB.RemoteProcedureCall(nodeA, "test", map[string]string{
		"x":  "42",
		"xs": "1,2,3,4,5",
	})

	s.Nil(err)
	s.NotNil(resChan)
	results := []m2cp.RpcResult{}
	for res := range resChan {
		resCast := res.(m2cp.RpcResult)
		results = append(results, resCast)
	}

	from, err := messages.NewAddress(s.ctp, "from-addr.foo")
	s.NoError(err)
	to, err := messages.NewAddress(s.ctp, "to-addr.foo")
	s.NoError(err)
	cmdMsg, _ := messages.NewCommandMessage(from, to, "Cmdname")
	responseMsg, err := messages.NewResponseMessage(cmdMsg)
	s.NoError(err)

	responseMsg.SetResults(results)
	json, err := messages.ToJson(responseMsg)
	s.NoError(err)
	fmt.Println(json)

}

func exampleCommand() m2cp.RpcCommand {
	cmd, err := rpc.NewRpcCommand(
		"Test",
		"desc",
		"desc_returns",
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeInt, "x", "int-param", false),
			rpc.NewRpcParameter(m2cp.TypeIntArray, "xs", "int-array-param", true),
		},
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeDouble, "d", "double-res", false),
			rpc.NewRpcParameter(m2cp.TypeInt, "y", "int-res", true),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			res := rpctypes.NewRpcResultSuccess("yeah")
			res.SetDouble("d", 3.141592653589793)
			res.SetInt("y", 42)
			return res
		})
	if err != nil {
		panic(err)
	}
	return cmd
}

func exampleCommandLoooongResponse() m2cp.RpcCommand {
	cmd, err := rpc.NewRpcCommand(
		"Logs",
		"desc",
		"desc_returns",
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeInt, "len", "int-param", false),
		},
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeDouble, "logs", "string", false),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {

			maxMessageSize := parameters.GetInt("len", 0)
			var logs string
			for len(logs) < maxMessageSize {
				logs = logs + "looong..." + fmt.Sprintf("%d", len(logs))
			}

			res := rpctypes.NewRpcResultSuccess("very long logs")
			res.SetString("logs", logs)
			return res
		})
	if err != nil {
		panic(err)
	}
	return cmd
}
