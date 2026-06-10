package protobuf

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/messages"
	"m2cp/networks"
	"m2cp/networks/amqp"
	"m2cp/tools"
	"math/rand"
	"strings"
	"testing"
	"unicode/utf8"
)

type OidTest struct {
	format m2cp.DataFormat
}

func NewOid17() m2cp.ProtobufOidParser {
	format, _ := messages.NewDataFormat(
		"mlpa_oid_17",
		messages.NewDataFieldString("test"))
	format.SetDataQueueMaxCount(0)
	return &OidTest{
		format: format,
	}
}

func (o *OidTest) GetOid() uint {
	return 0
}

func (o *OidTest) Parse(data []byte) ([]m2cp.DataRow, error) {

	row, err := o.format.NewRow(map[string]interface{}{
		"test": "test",
	})
	if err != nil {
		return nil, err
	}

	return []m2cp.DataRow{
		row,
	}, nil
}

// TestNewProtobufParserService doesn't do much - it just tests, that the constructor doesn't fail

func TestNewProtobufParserService(t *testing.T) {
	amqp.DebugSender = true
	amqp.DebugSubscriber = true

	ctp := contextplus.NewContextPlus()

	// we mock a m2cp-coap instance sending unparsed protobuf messages
	//ctp.LogInfo("Starting m2cp-coap mock")
	//m2cpCoapApp, err := newMockM2cpCoapApp(ctp)

	// we create a new protobuf parser service
	ctp.LogInfo("Starting protobuf parser service")
	con, err := networks.NewNetworkConnection(ctp)
	assert.NoError(t, err)
	assert.NotNil(t, con)

	svc, err := NewProtobufParserService(ctp, con, []m2cp.ProtobufOidParser{NewOid17()})
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	//// send mock data
	//ctp.LogInfo("Sending mock data")
	//assert.NoError(t, m2cpCoapApp.SendUnparsedProtobufMessage())
	//
	//ctp.Sleep(5 * time.Second)

	ctp.Cancel()
}

type mockM2cpCoapApp struct {
	ctp                m2cp.ContextPlus
	con                m2cp.NetworkConnection
	node               m2cp.Node
	sensorId           string
	oid                uint
	formatDataProtobuf m2cp.DataFormat
}

func newMockM2cpCoapApp(ctp m2cp.ContextPlus) (app *mockM2cpCoapApp, err error) {
	app = &mockM2cpCoapApp{
		ctp:      ctp,
		oid:      41,
		sensorId: tools.EncodeSensorId(425),
	}

	app.con, err = networks.NewNetworkConnectionWithOptions(ctp, m2cp.NetworkConnectionOptions{
		AppName: "m2cp-coap",
	})
	if err != nil {
		return nil, err
	}

	nodeName := fmt.Sprintf("sensor.%s.oid.%d.protobuf", app.sensorId, app.oid)
	app.node, err = app.con.NewNode(nodeName)
	if err != nil {
		return nil, err
	}

	fields := []m2cp.DataField{
		messages.NewDataFieldString("sid"),
		messages.NewDataFieldBinary("data"),
	}
	app.formatDataProtobuf, _ = messages.NewDataFormat("coap_data_protobuf", fields...)
	app.formatDataProtobuf.SetScope(m2cp.MessageScopeDevice)
	app.formatDataProtobuf.SetDataQueueMaxCount(0)

	app.con.AwaitReady()

	return app, nil
}

/*
func (o *mockM2cpCoapApp) SendUnparsedProtobufMessage() error {
	protoData := o.proto41Mock(uint64(123456))

	testUnparse := &DftDeviceProperties{}
	if err := proto.Unmarshal(protoData.Protobuf, testUnparse); err != nil {
		return fmt.Errorf("failed test-unmarshalling mock protobuf: " + err.Error())
	}

	row, err := o.formatDataProtobuf.NewRow(map[string]interface{}{
		"sid":  o.sensorId,
		"data": protoData.Protobuf,
	})
	if err != nil {
		return fmt.Errorf("failed creating data row: " + err.Error())
	}

	if err = o.node.EmitDataRow(row); err != nil {
		return fmt.Errorf("failed emitting data row: " + err.Error())
	}

	return nil
}

func (o *mockM2cpCoapApp) proto41Mock(timestamp uint64) WrappedProtobuf {
	deterministic := rand.New(rand.NewSource(int64(timestamp)))

	deviceProperties := &DftDeviceProperties{
		Timestamp:             &timestamp,
		DevName:               proto.String(randomString(deterministic)),
		DevType:               proto.String(randomString(deterministic)),
		DevModelType:          proto.String(randomString(deterministic)),
		DevSerialNumber:       proto.String(randomString(deterministic)),
		DevPartNumber:         proto.String(randomString(deterministic)),
		DevHardwareRevision:   proto.String(randomString(deterministic)),
		DevBuildNumber:        proto.String(randomString(deterministic)),
		DevBootloaderInfo:     proto.String(randomString(deterministic)),
		DevOsVersion:          proto.String(randomString(deterministic)),
		DevFirmwareVersion:    proto.String(randomString(deterministic)),
		DevActiveFirmware:     proto.String(randomString(deterministic)),
		DevApplicationVersion: proto.String(randomString(deterministic)),
		DevFactoryData:        randomBytes(deterministic),
		DevCustomData:         randomBytes(deterministic),
		DevTestData:           randomBytes(deterministic),
		DevConfigurationData:  randomBytes(deterministic),
	}

	data, err := proto.Marshal(deviceProperties)
	if err != nil {
		panic(err) // For mock, assume this never happens
	}

	return WrappedProtobuf{41, data}
}*/

func randomString(r *rand.Rand) string {
	// Extended character set including control characters, whitespaces, common symbols, and multibyte characters
	charset := "abc\n\r\t\x00\x1Fdefghi☃★♞\U0001F602jklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length := r.Intn(4096) + 1

	var output strings.Builder
	output.Grow(length)

	for i := 0; i < length; i++ {
		randomIndex := r.Intn(len(charset))
		runeValue, _ := utf8.DecodeRuneInString(charset[randomIndex:])
		output.WriteRune(runeValue)
	}
	return output.String()
}

func randomBytes(r *rand.Rand) []byte {
	n := r.Intn(1<<10) + 1
	bs := make([]byte, n)
	r.Read(bs) // Generates n random bytes
	return bs
}
