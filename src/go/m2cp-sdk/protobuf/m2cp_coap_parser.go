package protobuf

import (
	"encoding/json"
	"fmt"
	"m2cp"
	"m2cp/tools"
	"strings"
	"time"
)

// This file contains a framework for super simple creation of m2cp-coap-<project> parsing apps.
// Usage:
// Create a new protobuf parser service with NewProtobufParserService(..) and pass it a list of protobuf parsers.
// The created service will automatically subscribe to all topics of the form "sensor.*.oid.<oid>.protobuf" and parse the data using the provided protobuf parser.

type protobufParserService struct {
	ctp                m2cp.ContextPlus
	conn               m2cp.NetworkConnection
	node               m2cp.Node
	formatDataProtobuf m2cp.DataFormat
	forwarder          m2cp.RTDMessageForwarder
}

func NewProtobufParserService(ctpParent m2cp.ContextPlus, con m2cp.NetworkConnection, parsers []m2cp.ProtobufOidParser) (m2cp.ProtobufParserService, error) {
	if ctpParent == nil {
		return nil, fmt.Errorf("ctpParent is nil")
	}

	if con == nil {
		return nil, fmt.Errorf("con is nil")
	}

	if len(parsers) == 0 {
		return nil, fmt.Errorf("no parsers provided")
	}

	parseNode, err := con.NewNode("log.protobuf-parser")
	if err != nil {
		return nil, fmt.Errorf("creating parser log node: %v", err)
	}

	ctp := ctpParent.BranchWithName("pb-parser-svc")

	forwarder, err := NewRTDMessageForwarder(ctp, con, "protobuf-parser")
	if err != nil {
		return nil, fmt.Errorf("creating RTD message forwarder: %v", err)
	}

	svc := protobufParserService{
		ctp:       ctp,
		conn:      con,
		node:      parseNode,
		forwarder: forwarder,
	}

	parserOids := make([]uint, len(parsers))
	for i, parser := range parsers {
		RegisterParser(parser.GetOid(), parser)
		parserOids[i] = parser.GetOid()
	}

	if err := svc.registerProtobufTransformers(parserOids); err != nil {
		svc.ctp.Cancel()
		return nil, err
	}

	go svc.heartbeatWorker()

	con.AwaitReady()

	return &svc, nil
}

func (o *protobufParserService) heartbeatWorker() {
	go func() {
		for {
			select {
			case <-o.ctp.Done():
				return
			default:
				var statusJson []byte
				func() {
					heartbeatStats := o.forwarder.GetHeartBeatStats()
					statusJson, _ = json.Marshal(heartbeatStats)
				}()
				o.ctp.LogDebug(string(statusJson))
				o.emitLogSignal("status", string(statusJson), m2cp.SignalLevelInfo)
				o.ctp.Sleep(10 * time.Second)
			}
		}
	}()

}

func (o *protobufParserService) emitLogSignal(name, content, level string) {
	if err := o.node.EmitSignal(name, content, level); err != nil {
		o.ctp.LogError("failed emitting signal: " + err.Error())
		return
	}
}

func (o *protobufParserService) registerProtobufTransformers(oids []uint) error {

	for _, oid := range oids {
		err := o.subscribeDin1(oid)
		if err != nil {
			return err
		}
		err = o.subscribeDin2(oid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *protobufParserService) subscribeDin1(oid uint) error {
	parser, err := GetOidParser(oid)
	if err != nil {
		return err
	}
	topicDin1 := fmt.Sprintf("sensor.*.oid.%d.protobuf", oid)

	optionsDin1 := m2cp.SubscriptionOptions{
		PersistenceId: tools.StrPtr("parser-" + topicDin1),
	}
	if err := o.conn.SubscribeDataWithOptions([]string{topicDin1}, func(message m2cp.DataMessage) {
		// catch panic
		defer func() {
			if r := recover(); r != nil {
				o.ctp.LogError("protobuf transformer panic: %v", r)
			}
		}()

		o.forwarder.IncReceived()

		o.ctp.LogDebug("Received OID=%d data row: %v", oid, message)
		rows := message.GetRows()
		for _, row := range rows {
			data := row.GetFieldBinary("data")
			sensorId := row.GetFieldString("sid")
			if sensorId == nil {
				o.forwarder.IncFailed()
				o.ctp.LogError("missing sensorId in data row")
				continue
			}
			if parsed, err := parser.Parse(data); err != nil {
				o.forwarder.IncFailed()
				o.ctp.LogError("failed parsing protobuf data: %s", err)
				continue
			} else {
				o.forwarder.IncParsed()
				o.forwarder.EmitParsedData(m2cp.DIN1(*sensorId), oid, parsed)
			}
		}
	}, optionsDin1); err != nil {
		o.forwarder.IncFailed()
		return err
	}
	return nil
}

func (o *protobufParserService) subscribeDin2(oid uint) error {
	parser, err := GetOidParser(oid)
	if err != nil {
		return err
	}
	topicDin2 := fmt.Sprintf("oid.%d.protobuf", oid)
	optionsDin2 := m2cp.SubscriptionOptions{
		PersistenceId: tools.StrPtr("parser-" + topicDin2),
	}
	if err = o.conn.SubscribeDataWithOptions([]string{topicDin2}, func(message m2cp.DataMessage) {
		// catch panic
		defer func() {
			if r := recover(); r != nil {
				o.ctp.LogError("protobuf transformer panic: %v", r)
			}
		}()

		o.forwarder.IncReceived()

		origin := strings.Split(message.GetHeader().GetOrigin(), ".")
		osSerial := origin[len(origin)-1]

		o.ctp.LogDebug("Received OID=%d data row: %v", oid, message)
		rows := message.GetRows()
		for _, row := range rows {
			data := row.GetFieldBinary("data")

			if parsed, err := parser.Parse(data); err != nil {
				o.forwarder.IncFailed()
				o.ctp.LogError("failed parsing protobuf data: %s", err)
				continue
			} else {
				o.forwarder.IncParsed()
				o.forwarder.EmitParsedData(m2cp.DIN2(osSerial), oid, parsed)
			}
		}
	}, optionsDin2); err != nil {
		o.forwarder.IncFailed()
		return err
	}
	return nil
}
