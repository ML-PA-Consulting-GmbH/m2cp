package messages

import (
	"fmt"
	"os"
	"time"
)

const (
	exampleData0 = `{"Header":{"Type":"DATA","Topic":"data/sensor.3CIAAA.oid.104.parsed/","Timestamp":"1717824189.773194","Id":"62e2176a-9984-40b5-ba84-0e60f6cc7af7","TaskId":"28b6b21c-badd-4277-8742-6fe8a21bf849","Origin":"sensor.3CIAAA.oid.104.parsed.m2cp-coap-knorr.c60953c4-05d8-583a-ad92-7f44b3473d00"},"Body":{"FormatId":"mlpa_oid_104","Columns":[{"Name":"timestamp","Type":"uint64"},{"Name":"brakeMonitoringC1","Type":"uint32"},{"Name":"brakeMonitoringSensor","Type":"uint32"},{"Name":"brakeMonitoringT1","Type":"uint32"},{"Name":"brakeMonitoringWakeup","Type":"uint32"},{"Name":"brakeMonitoringA","Type":"uint32"},{"Name":"brakeMonitoringR","Type":"uint32"},{"Name":"brakeMonitoringCV","Type":"uint32"},{"Name":"brakeMonitoringHL","Type":"uint32"}],"Data":[{"T":"1717824.024","D":["1717824024","1263","11561","5874","3731","5987","6727","2988","4775"]}]},"Trace":[{"T":"1717824189777.1523","A":"sys-gw.c60953c4-05d8-583a-ad92-7f44b3473d00"}]}`
)

func (t *TestSuite) generateMockDataMessage(numRows int) []byte {

	var err error
	format, err := NewDataFormat("testFormat",
		NewDataFieldString("pressureValveA"),
		NewDataFieldString("pressureValveB"),
		NewDataFieldString("pressureValveC"),
		NewDataFieldString("temperatureRaw"),
		NewDataFieldString("fooReading"),
	)
	t.NoError(err)

	addr, err := NewAddressWithSubtopic(t.ctp, "foo.app.11111111-1111-1111-1111-111111111111", "bar")
	t.NotNil(addr)
	t.NoError(err)
	msg, err := NewDataMessage(addr, format)
	msg.GetHeader().SetTimestamp(1680075472.77354)
	msg.GetHeader().SetId("dfccbdf5-aa49-4529-af83-ffdda81cdee1")
	msg.GetHeader().SetTaskId("eb94490d-306b-4475-9e8b-5b470d669f80")

	for i := 0; i < numRows; i++ {
		values := map[string]interface{}{
			"pressureValveA": generateRandomString(10),
			"pressureValveB": generateRandomString(10),
			"pressureValveC": generateRandomString(10),
			"temperatureRaw": generateRandomString(10),
			"fooReading":     generateRandomString(10),
		}
		row, err := format.NewRow(values)
		t.NoError(err)
		timeC := time.Unix(1600000042, 3000)
		row.SetTime(timeC)

		msg.AddRow(row)
	}

	json, err := ToJson(msg)
	t.NoError(err)

	return json
}

func (t *TestSuite) TestPrintMockDataMessages() {
	const (
		numMessages    = 1
		rowsPerMessage = 10
	)

	out := ""
	for i := 0; i < numMessages; i++ {
		out += string(t.generateMockDataMessage(rowsPerMessage))
	}

	fname := fmt.Sprintf("/tmp/m2cp-mock-data_%dx%d.json", rowsPerMessage, numMessages)
	t.NoError(os.WriteFile(fname, []byte(out), 0644))
	fmt.Printf("Wrote %d messages with %d rows each to %s\n", numMessages, rowsPerMessage, fname)
}

func (t *TestSuite) TestProtobufCommand() {
	m, err := DataFromJson([]byte(exampleData0))
	t.NoError(err)

	pm, err := ToProtobuf(m)
	t.NoError(err)

	t.NoError(os.WriteFile("/tmp/m2cp-mock-proto", pm, 0644))

	//mParsed, err = FromProtobuf(pm)

	_ = pm

}
