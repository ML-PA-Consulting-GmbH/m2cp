package protobuf

import (
	"fmt"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/device"
	"m2cp/messages"
	"m2cp/networks"
	"m2cp/tools"
	"strings"
	"time"
)

func (t *TestSuite) TestNewRTDMessageForwarder_NilCTPParent() {
	conn, err := networks.NewNetworkConnection(t.ctp)
	t.NoError(err)
	_, err = NewRTDMessageForwarder(nil, conn, "app-name")
	t.Assert().Error(err)
}

func (t *TestSuite) TestNewRTDMessageForwarder_NilConn() {
	_, err := NewRTDMessageForwarder(t.ctp, nil, "app-name")
	t.Assert().Error(err)
}

func (t *TestSuite) TestNewRTDMessageForwarder_EmptyAppName() {
	conn, err := networks.NewNetworkConnection(t.ctp)
	t.NoError(err)
	_, err = NewRTDMessageForwarder(t.ctp, conn, "")
	t.Assert().Error(err)
}

func (t *TestSuite) TestNewRTDMessageForwarder_Success() {
	conn, err := networks.NewNetworkConnection(t.ctp)
	t.NoError(err)
	forwarder, err := NewRTDMessageForwarder(t.ctp, conn, "app-name")
	t.Assert().NoError(err)
	t.Assert().NotNil(forwarder)
}

func (t *TestSuite) TestRTDMessageForwarder_HeartBeatStats() {
	conn, err := networks.NewNetworkConnection(t.ctp)
	t.NoError(err)
	forwarder, err := NewRTDMessageForwarder(t.ctp, conn, "app-name")
	t.Assert().NoError(err)
	t.Assert().NotNil(forwarder)

	t.ctp.Sleep(1 * time.Second)
	uptime := forwarder.GetHeartBeatStats().Uptime
	t.Assert().GreaterOrEqual(uptime, uint32(1))
	t.Assert().LessOrEqual(uptime, uint32(2))

	forwarder.IncReceived()
	forwarder.IncParsed()
	forwarder.IncForwarded()
	forwarder.IncFailed()

	stats := forwarder.GetHeartBeatStats()
	t.Assert().Equal(uint32(1), stats.Received)
	t.Assert().Equal(uint32(1), stats.Parsed)
	t.Assert().Equal(uint32(1), stats.Forwarded)
	t.Assert().Equal(uint32(1), stats.Failed)

}

func (t *TestSuite) TestRTDMessageForwarder_EmitLogSignalDin1() {
	sensorId := "TESTSENSOR123"
	topic := "sensor.TESTSENSOR123.log"
	appName := "test-app"
	deviceName, err := device.GetName(t.ctp)
	t.Require().NoError(err)

	// Set up connection and forwarder
	conn, err := networks.NewNetworkConnectionWithOptions(t.ctp, m2cp.NetworkConnectionOptions{
		AppName: appName,
	})
	t.NoError(err)
	forwarder, err := NewRTDMessageForwarder(t.ctp, conn, appName)
	t.Assert().NoError(err)
	t.Assert().NotNil(forwarder)

	// Subscribe to log signal
	ctpSub := contextplus.NewContextPlus().SetModule("test-rtd-log-din1")
	messageReceivedCh := make(chan m2cp.SignalMessage)
	go func() {
		m, err := subscribeToSignalFromTopic(ctpSub, topic)
		t.Assert().NoError(err)
		messageReceivedCh <- m
	}()

	t.ctp.Sleep(1 * time.Second) // give some time to subscribe

	// Emit log signal and wait for reception
	forwarder.EmitLogSignal(m2cp.DIN1(sensorId), "test-log", "This is a test log message", m2cp.SignalLevelInfo)
	m := <-messageReceivedCh

	// Validate received message
	t.Require().NotNil(m)
	t.Assert().Equal("test-log", m.GetName())
	t.Assert().Equal("This is a test log message", m.GetContent())
	expectedOrigin := fmt.Sprintf("%s.%s.%s", topic, appName, deviceName)
	t.Assert().Equal(expectedOrigin, m.GetHeader().GetOrigin())
}

func (t *TestSuite) TestRTDMessageForwarder_EmitLogSignalDin2() {
	osSerial := "ab-cd-ef"
	topic := "log"
	appName := "test-app"

	// Set up connection and forwarder
	conn, err := networks.NewNetworkConnectionWithOptions(t.ctp, m2cp.NetworkConnectionOptions{
		AppName: appName,
	})
	t.NoError(err)
	forwarder, err := NewRTDMessageForwarder(t.ctp, conn, appName)
	t.Assert().NoError(err)
	t.Assert().NotNil(forwarder)

	// Subscribe to log signal
	ctpSub := contextplus.NewContextPlus().SetModule("test-rtd-log-din2")
	messageReceivedCh := make(chan m2cp.SignalMessage)
	go func() {
		m, err := subscribeToSignalFromTopic(ctpSub, topic)
		t.Assert().NoError(err)
		messageReceivedCh <- m
	}()

	t.ctp.Sleep(1 * time.Second) // give some time to subscribe

	// Emit log signal and wait for reception
	forwarder.EmitLogSignal(m2cp.DIN2(osSerial), "test-log", "This is a test log message", m2cp.SignalLevelInfo)
	m := <-messageReceivedCh

	// Validate received message
	t.Require().NotNil(m)
	t.Assert().Equal("test-log", m.GetName())
	t.Assert().Equal("This is a test log message", m.GetContent())
	expectedOrigin := fmt.Sprintf("%s.%s.%s", topic, appName, osSerial)
	t.Assert().Equal(expectedOrigin, m.GetHeader().GetOrigin())
}

func (t *TestSuite) TestRTDMessageForwarder_EmitParsedDataDin1() {
	sensorId := "TESTSENSOR123"
	oid := uint(123)
	topic := fmt.Sprintf("sensor.%s.oid.%d.parsed", sensorId, oid)
	appName := "test-app"
	deviceName, err := device.GetName(t.ctp)
	t.Require().NoError(err)

	// Set up connection and forwarder
	conn, err := networks.NewNetworkConnectionWithOptions(t.ctp, m2cp.NetworkConnectionOptions{
		AppName: appName,
	})
	t.NoError(err)
	forwarder, err := NewRTDMessageForwarder(t.ctp, conn, appName)
	t.Assert().NoError(err)
	t.Assert().NotNil(forwarder)

	// Subscribe to data
	ctpSub := contextplus.NewContextPlus().SetModule("test-rtd-parsed-din1")
	messageReceivedCh := make(chan m2cp.DataMessage)
	go func() {
		m, err := subscribeToDataFromTopic(ctpSub, topic)
		t.Assert().NoError(err)
		messageReceivedCh <- m
	}()

	t.ctp.Sleep(1 * time.Second) // give some time to subscribe

	// Emit parsed data and wait for reception
	format := createTestParsedDataFormat(t)
	testRow, err := format.NewRow(map[string]interface{}{
		"temperature": 25.5,
		"humidity":    60.0,
	})
	t.Require().NoError(err)
	forwarder.EmitParsedData(m2cp.DIN1(sensorId), oid, []m2cp.DataRow{testRow})
	m := <-messageReceivedCh

	// Validate received message
	t.Require().NotNil(m)
	rows := m.GetRows()
	t.Assert().Len(rows, 1)
	t.Assert().Equal(25.5, *rows[0].GetFieldDouble("temperature"))
	t.Assert().Equal(60.0, *rows[0].GetFieldDouble("humidity"))
	expectedOrigin := fmt.Sprintf("%s.%s.%s", topic, appName, deviceName)
	t.Assert().Equal(expectedOrigin, m.GetHeader().GetOrigin())
}

func (t *TestSuite) TestRTDMessageForwarder_EmitParsedDataDin2() {
	osSerial := "ab-cd-ef"
	oid := uint(123)
	topic := fmt.Sprintf("oid.%d.parsed", oid)
	appName := "test-app"

	// Set up connection and forwarder
	conn, err := networks.NewNetworkConnectionWithOptions(t.ctp, m2cp.NetworkConnectionOptions{
		AppName: appName,
	})
	t.NoError(err)
	forwarder, err := NewRTDMessageForwarder(t.ctp, conn, appName)
	t.Assert().NoError(err)
	t.Assert().NotNil(forwarder)

	// Subscribe to data
	ctpSub := contextplus.NewContextPlus().SetModule("test-rtd-parsed-din2")
	messageReceivedCh := make(chan m2cp.DataMessage)
	go func() {
		m, err := subscribeToDataFromTopic(ctpSub, topic)
		t.Assert().NoError(err)
		messageReceivedCh <- m
	}()

	t.ctp.Sleep(1 * time.Second) // give some time to subscribe

	// Emit parsed data and wait for reception
	format := createTestParsedDataFormat(t)
	testRow, err := format.NewRow(map[string]interface{}{
		"temperature": 25.5,
		"humidity":    60.0,
	})
	t.Require().NoError(err)
	forwarder.EmitParsedData(m2cp.DIN2(osSerial), oid, []m2cp.DataRow{testRow})
	m := <-messageReceivedCh

	// Validate received message
	t.Require().NotNil(m)
	rows := m.GetRows()
	t.Assert().Len(rows, 1)
	t.Assert().Equal(25.5, *rows[0].GetFieldDouble("temperature"))
	t.Assert().Equal(60.0, *rows[0].GetFieldDouble("humidity"))
	expectedOrigin := fmt.Sprintf("%s.%s.%s", topic, appName, osSerial)
	t.Assert().Equal(expectedOrigin, m.GetHeader().GetOrigin())
}

func (t *TestSuite) TestRTDMessageForwarder_EmitUnparsedDataDin1() {
	sensorId := "TESTSENSOR123"
	oid := uint(456)
	topic := fmt.Sprintf("sensor.%s.oid.%d.protobuf", sensorId, oid)
	appName := "test-app"
	deviceName, err := device.GetName(t.ctp)
	t.Require().NoError(err)

	// Set up connection and forwarder
	conn, err := networks.NewNetworkConnectionWithOptions(t.ctp, m2cp.NetworkConnectionOptions{
		AppName: appName,
	})
	t.NoError(err)
	forwarder, err := NewRTDMessageForwarder(t.ctp, conn, appName)
	t.Assert().NoError(err)
	t.Assert().NotNil(forwarder)

	// Subscribe to data
	ctpSub := contextplus.NewContextPlus().SetModule("test-rtd-unparsed-din1")
	messageReceivedCh := make(chan m2cp.DataMessage)
	go func() {
		m, err := subscribeToDataFromTopic(ctpSub, topic)
		t.Assert().NoError(err)
		messageReceivedCh <- m
	}()

	t.ctp.Sleep(1 * time.Second) // give some time to subscribe

	// Emit unparsed data and wait for reception
	format := createTestUnparsedDataFormatDin1(t)
	testRow, err := format.NewRow(map[string]interface{}{
		"sid":  sensorId,
		"data": []byte{0x01, 0x02, 0x03, 0x04},
	})
	t.Require().NoError(err)
	forwarder.ForwardUnparsedDataMessage(m2cp.DIN1(sensorId), oid, []m2cp.DataRow{testRow})
	m := <-messageReceivedCh

	// Validate received message
	t.Require().NotNil(m)
	rows := m.GetRows()
	t.Assert().Len(rows, 1)
	expectedOrigin := fmt.Sprintf("%s.%s.%s", topic, appName, deviceName)
	t.Assert().Equal(expectedOrigin, m.GetHeader().GetOrigin())
}

func (t *TestSuite) TestRTDMessageForwarder_EmitUnparsedDataDin2() {
	osSerial := "ab-cd-ef"
	oid := uint(456)
	topic := fmt.Sprintf("oid.%d.protobuf", oid)
	appName := "test-app"

	// Set up connection and forwarder
	conn, err := networks.NewNetworkConnectionWithOptions(t.ctp, m2cp.NetworkConnectionOptions{
		AppName: appName,
	})
	t.NoError(err)
	forwarder, err := NewRTDMessageForwarder(t.ctp, conn, appName)
	t.Assert().NoError(err)
	t.Assert().NotNil(forwarder)

	// Subscribe to data
	ctpSub := contextplus.NewContextPlus().SetModule("test-rtd-unparsed-din2")
	messageReceivedCh := make(chan m2cp.DataMessage)
	go func() {
		m, err := subscribeToDataFromTopic(ctpSub, topic)
		t.Assert().NoError(err)
		messageReceivedCh <- m
	}()

	t.ctp.Sleep(1 * time.Second) // give some time to subscribe

	// Emit unparsed data and wait for reception
	format := createTestUnparsedDataFormatDin2(t)
	testRow, err := format.NewRow(map[string]interface{}{
		"data": []byte{0x01, 0x02, 0x03, 0x04},
	})
	t.Require().NoError(err)
	forwarder.ForwardUnparsedDataMessage(m2cp.DIN2(osSerial), oid, []m2cp.DataRow{testRow})
	m := <-messageReceivedCh

	// Validate received message
	t.Require().NotNil(m)
	rows := m.GetRows()
	t.Assert().Len(rows, 1)
	expectedOrigin := fmt.Sprintf("%s.%s.%s", topic, appName, osSerial)
	t.Assert().Equal(expectedOrigin, m.GetHeader().GetOrigin())
}

func subscribeToDataFromTopic(ctp m2cp.ContextPlus, topic string) (m2cp.DataMessage, error) {

	conn, err := networks.NewNetworkConnectionWithOptions(ctp, m2cp.NetworkConnectionOptions{
		AppName: "test-listener",
	})
	if err != nil {
		return nil, err
	}

	optionsDin1 := m2cp.SubscriptionOptions{
		PersistenceId: tools.StrPtr("test-subscribe-" + topic),
		Context:       ctp,
	}

	var message m2cp.DataMessage
	if err := conn.SubscribeDataWithOptions([]string{topic}, func(msg m2cp.DataMessage) {
		// catch panic
		defer func() {
			if r := recover(); r != nil {
				ctp.LogError("protobuf transformer panic: %v", r)
			}
		}()

		ctp.LogDebug("Received data row: %v", msg)
		message = msg
		// stop after first message
		ctp.Cancel()

	}, optionsDin1); err != nil {
		return nil, err
	}

	select {
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for message on topic %s", topic)
	case <-ctp.Done():
		if message != nil {
			return message, nil
		}
		return nil, fmt.Errorf("context cancelled while waiting for message on topic %s", topic)
	}

}

func subscribeToSignalFromTopic(ctp m2cp.ContextPlus, topic string) (m2cp.SignalMessage, error) {

	conn, err := networks.NewNetworkConnectionWithOptions(ctp, m2cp.NetworkConnectionOptions{
		AppName: "test-listener",
	})
	if err != nil {
		return nil, err
	}

	optionsDin1 := m2cp.SubscriptionOptions{
		PersistenceId: tools.StrPtr("test-subscribe-signal-" + topic),
		Context:       ctp,
	}

	var message m2cp.SignalMessage
	if err := conn.SubscribeSignalsWithOptions([]string{topic}, func(msg m2cp.SignalMessage) {
		// catch panic
		defer func() {
			if r := recover(); r != nil {
				ctp.LogError("protobuf transformer panic: %v", r)
			}
		}()

		if strings.Contains(msg.GetName(), "hello!") {
			return
		}

		ctp.LogDebug("Received signal: %v", msg)
		message = msg
		// stop after first message
		ctp.Cancel()

	}, optionsDin1); err != nil {
		return nil, err
	}

	select {
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for signal on topic %s", topic)
	case <-ctp.Done():
		if message != nil {
			return message, nil
		}
		return nil, fmt.Errorf("context cancelled while waiting for signal on topic %s", topic)
	}

}

// createTestParsedDataFormat creates a data format for testing parsed data messages
func createTestParsedDataFormat(t *TestSuite) m2cp.DataFormat {
	fields := []m2cp.DataField{
		messages.NewDataFieldDouble("temperature"),
		messages.NewDataFieldDouble("humidity"),
	}
	format, err := messages.NewDataFormat("test_parsed_format", fields...)
	t.Require().NoError(err)
	format.SetScope(m2cp.MessageScopeDevice)
	format.SetDataQueueMaxCount(0)
	return format
}

// createTestUnparsedDataFormat creates a data format for testing unparsed/protobuf data messages
func createTestUnparsedDataFormatDin1(t *TestSuite) m2cp.DataFormat {
	fields := []m2cp.DataField{
		messages.NewDataFieldString("sid"),
		messages.NewDataFieldBinary("data"),
	}
	format, err := messages.NewDataFormat("test_protobuf_format", fields...)
	t.Require().NoError(err)
	format.SetScope(m2cp.MessageScopeDevice)
	format.SetDataQueueMaxCount(0)
	return format
}

// createTestUnparsedDataFormat creates a data format for testing unparsed/protobuf data messages
func createTestUnparsedDataFormatDin2(t *TestSuite) m2cp.DataFormat {
	fields := []m2cp.DataField{
		messages.NewDataFieldBinary("data"),
	}
	format, err := messages.NewDataFormat("test_protobuf_format", fields...)
	t.Require().NoError(err)
	format.SetScope(m2cp.MessageScopeDevice)
	format.SetDataQueueMaxCount(0)
	return format
}
