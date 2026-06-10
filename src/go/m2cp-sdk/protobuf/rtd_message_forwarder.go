package protobuf

import (
	"fmt"
	"m2cp"
	"m2cp/networks"
	"sync"
	"time"
)

type rtdMessageForwarder struct {
	ctp                m2cp.ContextPlus
	appName            string
	heartbeatStats     m2cp.HeartBeatStats
	heartbeatStatsLock sync.RWMutex

	mainConnection        m2cp.NetworkConnection
	forwardingConnections map[string]m2cp.NetworkConnection
	nodes                 map[string]m2cp.Node
	messagingLock         sync.Mutex
}

func NewRTDMessageForwarder(ctpParent m2cp.ContextPlus, conn m2cp.NetworkConnection, appName string) (m2cp.RTDMessageForwarder, error) {
	if ctpParent == nil {
		return nil, fmt.Errorf("ctpParent is nil")
	}

	if conn == nil {
		return nil, fmt.Errorf("conn is nil")
	}

	if appName == "" {
		return nil, fmt.Errorf("appName is empty")
	}

	forwarder := &rtdMessageForwarder{
		ctp:     ctpParent.BranchWithName("rtd-message-forwarder"),
		appName: appName,
		heartbeatStats: m2cp.HeartBeatStats{
			Started: time.Now(),
		},
		heartbeatStatsLock: sync.RWMutex{},

		mainConnection:        conn,
		forwardingConnections: make(map[string]m2cp.NetworkConnection),
		nodes:                 make(map[string]m2cp.Node),
		messagingLock:         sync.Mutex{},
	}

	return forwarder, nil
}

func (o *rtdMessageForwarder) IncReceived() {
	o.heartbeatStatsLock.Lock()
	defer o.heartbeatStatsLock.Unlock()
	o.heartbeatStats.Received++
}

func (o *rtdMessageForwarder) IncParsed() {
	o.heartbeatStatsLock.Lock()
	defer o.heartbeatStatsLock.Unlock()
	o.heartbeatStats.Parsed++
}

func (o *rtdMessageForwarder) IncForwarded() {
	o.heartbeatStatsLock.Lock()
	defer o.heartbeatStatsLock.Unlock()
	o.heartbeatStats.Forwarded++
}
func (o *rtdMessageForwarder) IncFailed() {
	o.heartbeatStatsLock.Lock()
	defer o.heartbeatStatsLock.Unlock()
	o.heartbeatStats.Failed++
}

func (o *rtdMessageForwarder) GetHeartBeatStats() m2cp.HeartBeatStats {
	o.heartbeatStatsLock.Lock()
	defer o.heartbeatStatsLock.Unlock()
	o.heartbeatStats.Uptime = uint32(time.Since(o.heartbeatStats.Started).Seconds())
	return o.heartbeatStats

}

func (o *rtdMessageForwarder) EmitLogSignal(standard m2cp.RTDDataStandard, name, content, level string) {
	if standard.Standard == m2cp.StandardDIN1 && standard.HWSerial != "" {
		o.emitLogSignalDin1(standard.HWSerial, name, content, level)
	} else if standard.Standard == m2cp.StandardDIN2 && standard.OSSerial != "" {
		o.emitLogSignalDin2(standard.OSSerial, name, content, level)
	} else {
		o.IncFailed()
		o.ctp.LogError("EmitLogSignal was called without specifying din/1 or din/2 correctly")
	}
}

func (o *rtdMessageForwarder) emitLogSignalDin1(sensorId, name, content, level string) {
	node, err := o.getNodeFromParserConnection(fmt.Sprintf("sensor.%s.log", sensorId))
	if err != nil {
		o.ctp.LogError("failed getting log node: " + err.Error())
		return
	}
	if err = node.EmitSignal(name, content, level); err != nil {
		o.ctp.LogError("failed emitting signal: " + err.Error())
		return
	}
}
func (o *rtdMessageForwarder) emitLogSignalDin2(osSerial, name, content, level string) {

	node, err := o.getRTDNode(osSerial, o.appName, "log")
	if err != nil {
		o.ctp.LogError("failed getting log node: " + err.Error())
		return
	}
	if err = node.EmitSignal(name, content, level); err != nil {
		o.ctp.LogError("failed emitting signal: " + err.Error())
		return
	}
}

func (o *rtdMessageForwarder) EmitParsedData(standard m2cp.RTDDataStandard, oid uint, rows []m2cp.DataRow) {
	if standard.Standard == m2cp.StandardDIN1 && standard.HWSerial != "" {
		o.emitParsedDataRowsDin1(standard.HWSerial, oid, rows)
	} else if standard.Standard == m2cp.StandardDIN2 && standard.OSSerial != "" {
		o.emitParsedDataRowsDin2(standard.OSSerial, oid, rows)
	} else {
		o.IncFailed()
		o.ctp.LogError("EmitParsedData was called without specifying din/1 or din/2 correctly")
	}
}

func (o *rtdMessageForwarder) emitParsedDataRowsDin2(osSerial string, oid uint, rows []m2cp.DataRow) {
	node, err := o.getRTDNode(osSerial, o.appName, fmt.Sprintf("oid.%d.parsed", oid))
	if err != nil {
		o.ctp.LogError("failed getting data node: " + err.Error())
		return
	}

	if err = node.EmitDataRows(rows); err != nil {
		o.ctp.LogError("failed emitting data row: " + err.Error())
		return
	}
}

func (o *rtdMessageForwarder) emitParsedDataRowsDin1(sensorId string, oid uint, rows []m2cp.DataRow) {
	node, err := o.getNodeFromParserConnection(fmt.Sprintf("sensor.%s.oid.%d.parsed", sensorId, oid))
	if err != nil {
		o.ctp.LogError("failed getting data node: " + err.Error())
		return
	}

	if err = node.EmitDataRows(rows); err != nil {
		o.ctp.LogError("failed emitting data row: " + err.Error())
		return
	}
}

func (o *rtdMessageForwarder) ForwardUnparsedDataMessage(standard m2cp.RTDDataStandard, oid uint, rows []m2cp.DataRow) {
	if standard.Standard == m2cp.StandardDIN1 && standard.HWSerial != "" {
		o.emitUnparsedDataRowsDin1(standard.HWSerial, oid, rows)
	} else if standard.Standard == m2cp.StandardDIN2 && standard.OSSerial != "" {
		o.emitUnparsedDataRowsDin2(standard.OSSerial, oid, rows)
	} else {
		o.IncFailed()
		o.ctp.LogError("ForwardUnparsedDataMessage was called without specifying din/1 or din/2 correctly")
	}
}

func (o *rtdMessageForwarder) emitUnparsedDataRowsDin2(osSerial string, oid uint, rows []m2cp.DataRow) {
	node, err := o.getRTDNode(osSerial, o.appName, fmt.Sprintf("oid.%d.protobuf", oid))
	if err != nil {
		o.ctp.LogError("failed getting data node: " + err.Error())
		return
	}

	if err = node.EmitDataRows(rows); err != nil {
		o.ctp.LogError("failed emitting data row: " + err.Error())
		return
	}
}

func (o *rtdMessageForwarder) emitUnparsedDataRowsDin1(sensorId string, oid uint, rows []m2cp.DataRow) {
	node, err := o.getNodeFromParserConnection(fmt.Sprintf("sensor.%s.oid.%d.protobuf", sensorId, oid))
	if err != nil {
		o.ctp.LogError("failed getting data node: " + err.Error())
		return
	}

	if err = node.EmitDataRows(rows); err != nil {
		o.ctp.LogError("failed emitting data row: " + err.Error())
		return
	}
}

func (o *rtdMessageForwarder) getRTDNode(osSerial, appName, feature string) (m2cp.Node, error) {
	o.messagingLock.Lock()
	defer o.messagingLock.Unlock()

	if osSerial == "" && appName == "" {
		return o.getNodeFromParserConnection(feature)
	}
	connKey := osSerial + "." + appName
	nodeKey := connKey + "." + feature

	if node, ok := o.nodes[nodeKey]; ok {
		return node, nil
	}

	conn, ok := o.forwardingConnections[connKey]
	if !ok {
		var err error
		conn, err = networks.NewNetworkConnectionWithOptions(o.ctp, m2cp.NetworkConnectionOptions{
			AppName:    appName,
			DeviceName: osSerial,
		})
		if err != nil {
			return nil, fmt.Errorf("creating network connection: %v", err)
		}
		o.forwardingConnections[connKey] = conn
	}
	node, err := conn.NewNode(feature)
	if err != nil {
		return nil, err
	}
	o.nodes[nodeKey] = node
	return node, nil
}

func (o *rtdMessageForwarder) getNodeFromParserConnection(feature string) (m2cp.Node, error) {
	if node, ok := o.nodes[feature]; ok {
		return node, nil
	}
	node, err := o.mainConnection.NewNode(feature)
	if err != nil {
		return nil, err
	}
	o.nodes[feature] = node
	return node, nil
}
