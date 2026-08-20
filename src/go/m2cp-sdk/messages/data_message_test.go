package messages

import (
	"encoding/base64"
	"fmt"
	"m2cp"
	"m2cp/tools"
	"strings"
	"time"
)

func (t *TestSuite) TestDataMessageTimePrecision() {
	var err error
	format, err := NewDataFormat("testFormat", NewDataFieldDatetime("t"))
	timeA := time.Unix(1600000000, 1)
	valuesA := map[string]interface{}{
		"t": timeA,
	}
	rowA, err := format.NewRow(valuesA)
	t.NoError(err)
	t.Equal(timeA.UTC(), *rowA.GetFieldDatetime("t"))

	format.SetScope(m2cp.MessageScopeDevice)

	addr, err := NewAddressWithSubtopic(t.ctp, "foo.app.11111111-1111-1111-1111-111111111111", "bar")
	t.NotNil(addr)
	t.NoError(err)
	msg, err := NewDataMessage(addr, format)
	t.NoError(err)

	t.Equal(m2cp.MessageScopeDevice, msg.GetFormat().GetScope())

	msg.AddRow(rowA)
	json, err := ToJson(msg)
	t.NoError(err)

	binary, _ := ToBinary(msg, m2cp.MessageSerializerJsonDeflate)
	// base64 encode
	fmt.Println(base64.StdEncoding.EncodeToString(binary))

	parsed, err := DataFromJson(json)
	t.NoError(err)
	res := parsed.GetRows()[0].GetFieldDatetime("t")
	t.Equal(timeA.UTC(), *res)

	t.Equal(m2cp.MessageScopeDevice, parsed.GetHeader().GetScope())
	t.Equal(m2cp.MessageScopeDevice, parsed.GetFormat().GetScope())
}

func (t *TestSuite) TestBuildConvertAndParseDataMessage() {
	var err error
	format, err := NewDataFormat("testFormat",
		NewDataFieldBool("good"),
		NewDataFieldInt("x"),
		NewDataFieldIntList("xs"),
		NewDataFieldUint32("u32"),
		NewDataFieldUint32List("u32s"),
		NewDataFieldUint64("u64"),
		NewDataFieldUint64List("u64s"),
		NewDataFieldDouble("d"),
		NewDataFieldDoubleList("ds"),
		NewDataFieldString("s"),
		NewDataFieldBinary("b"),
		NewDataFieldDatetime("t"),
	)
	t.NoError(err)

	timeA := time.Unix(1600000000, 1)
	valuesA := map[string]interface{}{
		"good": true,
		"x":    42,
		"xs":   []int{1, 2, 3},
		"u32":  uint32(42),
		"u32s": []uint32{1, 2, 3},
		"u64":  uint64(42),
		"u64s": []uint64{1, 2, 3},
		"d":    1.5,
		"ds":   []float64{1.1, 1.2, 1.3},
		"s":    "foo",
		"b":    []byte{0, 1, 2, 3, 4},
		"t":    timeA,
	}
	rowA, err := format.NewRow(valuesA)
	t.NoError(err)
	timeC := time.Unix(1600000042, 3000)
	rowA.SetTime(timeC)

	timeB := time.Unix(1600000000, 2000)
	valuesB := map[string]interface{}{
		"good": false,
		"x":    1,
		"xs":   []int{1, 2, 3},
		"u32":  uint32(1),
		"u32s": []uint32{1, 2, 3},
		"u64":  uint64(1),
		"u64s": []uint64{1, 2, 3},
		"d":    1.5,
		"ds":   []float64{1.21, 441.2, -551.3},
		"s":    "hello",
		"b":    []byte{42, 17, 12, 34, 124},
		"t":    timeB,
	}
	rowB, err := format.NewRow(valuesB)
	t.NoError(err)
	timeD := time.Unix(1600000043, 4000)
	rowB.SetTime(timeD)

	addr, err := NewAddressWithSubtopic(t.ctp, "foo.bar.local", "bar")
	t.NoError(err)
	msg, err := NewDataMessage(addr, format)
	msg.GetHeader().SetTimestamp(1680075472.77354)
	msg.GetHeader().SetId("dfccbdf5-aa49-4529-af83-ffdda81cdee1")
	msg.GetHeader().SetTaskId("eb94490d-306b-4475-9e8b-5b470d669f80")

	t.NoError(err)
	t.Equal("testFormat", msg.GetFormat().GetFormatId())
	msg.AddRow(rowA)
	msg.AddRows([]m2cp.DataRow{rowA, rowB})
	json, err := ToJson(msg)
	t.Equal(`{"Header":{"Type":"DATA","Topic":"data/bar","Scope":"GLOBAL","Timestamp":"1680075472.77354","Id":"dfccbdf5-aa49-4529-af83-ffdda81cdee1","TaskId":"eb94490d-306b-4475-9e8b-5b470d669f80","Origin":"`+addr.GetAddress()+`"},"Body":{"FormatId":"testFormat","Columns":[{"Name":"good","Type":"bool"},{"Name":"x","Type":"int"},{"Name":"xs","Type":"int[]"},{"Name":"u32","Type":"uint32"},{"Name":"u32s","Type":"uint32[]"},{"Name":"u64","Type":"uint64"},{"Name":"u64s","Type":"uint64[]"},{"Name":"d","Type":"double"},{"Name":"ds","Type":"double[]"},{"Name":"s","Type":"string"},{"Name":"b","Type":"binary"},{"Name":"t","Type":"datetime"}],"Data":[{"T":"1600000042.000003","D":["true","42","1,2,3","42","1,2,3","42","1,2,3","1.5","1.1,1.2,1.3","foo","AAECAwQ=","2020-09-13T12:26:40.000000001Z"]},{"T":"1600000042.000003","D":["true","42","1,2,3","42","1,2,3","42","1,2,3","1.5","1.1,1.2,1.3","foo","AAECAwQ=","2020-09-13T12:26:40.000000001Z"]},{"T":"1600000043.000004","D":["false","1","1,2,3","1","1,2,3","1","1,2,3","1.5","1.21,441.2,-551.3","hello","KhEMInw=","2020-09-13T12:26:40.000002Z"]}]}}`,
		string(json))
	t.NoError(err)
	parsed, err := DataFromJson(json)
	t.NoError(err)
	t.Equal("testFormat", parsed.GetFormat().GetFormatId())
	rows := parsed.GetRows()
	row := rows[2]
	t.Equal(false, *row.GetFieldBool("good"))
	t.Equal(1, *row.GetFieldInt("x"))
	xs := row.GetFieldIntArray("xs")
	t.NotNil(xs)
	t.Equal(1, xs[0])
	t.Equal(2, xs[1])
	t.Equal(3, xs[2])

	t.Equal(uint32(1), *row.GetFieldUint32("u32"))
	u32s := row.GetFieldUint32Array("u32s")
	t.NotNil(u32s)
	t.Equal(uint32(1), u32s[0])
	t.Equal(uint32(2), u32s[1])
	t.Equal(uint32(3), u32s[2])

	t.Equal(uint64(1), *row.GetFieldUint64("u64"))
	u64s := row.GetFieldUint64Array("u64s")
	t.NotNil(u64s)
	t.Equal(uint64(1), u64s[0])
	t.Equal(uint64(2), u64s[1])
	t.Equal(uint64(3), u64s[2])

	t.Equal(1.5, *row.GetFieldDouble("d"))
	ds := row.GetFieldDoubleArray("ds")
	t.NotNil(ds)
	t.Equal(1.21, ds[0])
	t.Equal(441.2, ds[1])
	t.Equal(-551.3, ds[2])
	t.Equal("hello", *row.GetFieldString("s"))
	t.Equal([]byte{42, 17, 12, 34, 124}, row.GetFieldBinary("b"))
	t.True(tools.TimesSimilar(timeD.UTC(), row.GetTime(), 1*time.Millisecond))
}

// https://dev.azure.com/ml-pa/M2CP/_workitems/edit/15725
func (t *TestSuite) TestDataMessage_15725_ParseOID32_withOptionalFields() {
	format, err := NewDataFormat(
		"mlpa_oid_23",
		NewDataFieldUint64("t"),
		NewDataFieldString("fwType"),
		NewDataFieldString("fwVersion"),
		NewDataFieldUint32("fwRevision"),
		NewDataFieldUint32("cfgIdentifier"),
		NewDataFieldUint32("appState"),
	)
	t.Require().NoError(err)

	testCases := []struct {
		name             string
		t                uint64
		fwType           string
		fwVersion        string
		fwRevision       *uint32
		cfgIdentifier    *uint32
		appState         *uint32
		expectedDataJson string
	}{
		{"All fields present",
			1769934987, "fw_type", "fw_version", Ptr(uint32(4)), Ptr(uint32(5)), Ptr(uint32(6)),
			`"1769934987","fw_type","fw_version","4","5","6"`},
		{"Optional fields nil",
			1600000000, "typeB", "versionB", nil, nil, nil,
			`"1600000000","typeB","versionB",null,null,null`},
		{"Some optional fields nil",
			1769934993, "fw_type", "fw_version", nil, nil, Ptr(uint32(6)),
			`"1769934993","fw_type","fw_version",null,null,"6"`},
		{"Default 0 values for optional uint32 fields",
			1600000000, "typeD", "versionD", Ptr(uint32(0)), Ptr(uint32(0)), Ptr(uint32(0)),
			`"1600000000","typeD","versionD","0","0","0"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func() {
			values := map[string]interface{}{
				"t":             tc.t,
				"fwType":        tc.fwType,
				"fwVersion":     tc.fwVersion,
				"fwRevision":    tc.fwRevision,
				"cfgIdentifier": tc.cfgIdentifier,
				"appState":      tc.appState,
			}
			rowToSend, err := format.NewRow(values)
			t.Require().NoError(err)
			rowTime := time.Unix(1600000043, 4000)
			rowToSend.SetTime(rowTime)

			addr, err := NewAddressWithSubtopic(t.ctp, "foo.bar.local", "status")
			t.Require().NoError(err)
			msg, err := NewDataMessage(addr, format)
			t.Require().NoError(err)
			msg.GetHeader().SetTimestamp(1680075472.77354)
			msg.GetHeader().SetId("99035d37-0455-439c-b6a7-146565d9951e")
			msg.GetHeader().SetTaskId("eb94490d-306b-4475-9e8b-5b470d669f80")
			t.Equal("mlpa_oid_23", msg.GetFormat().GetFormatId())

			msg.AddRow(rowToSend)
			json, err := ToJson(msg)
			t.JSONEq(fmt.Sprintf(`{
"Header": {
  "Type":"DATA",
  "Topic":"data/status",
  "Scope":"GLOBAL",
  "Timestamp":"1680075472.77354",
  "Id":"99035d37-0455-439c-b6a7-146565d9951e",
  "TaskId":"eb94490d-306b-4475-9e8b-5b470d669f80",
  "Origin":"%s"
},"Body": {
  "FormatId":"mlpa_oid_23",
  "Columns":[
  {"Name":"t","Type":"uint64"},
  {"Name":"fwType","Type":"string"},
  {"Name":"fwVersion","Type":"string"},
  {"Name":"fwRevision","Type":"uint32"},
  {"Name":"cfgIdentifier","Type":"uint32"},
  {"Name":"appState","Type":"uint32"}],
  "Data":[
  {"T":"1600000043.000004","D":[%s]}
 ]}
}`, addr.GetAddress(), tc.expectedDataJson), string(json))
			t.Require().NoError(err)

			parsed, err := DataFromJson(json)
			t.Require().NoError(err)
			t.Equal("mlpa_oid_23", parsed.GetFormat().GetFormatId())
			rows := parsed.GetRows()
			row := rows[0]

			t.Equal(tc.t, *row.GetFieldUint64("t"))
			t.Equal(tc.fwType, *row.GetFieldString("fwType"))
			t.Equal(tc.fwVersion, *row.GetFieldString("fwVersion"))
			if tc.fwRevision != nil {
				t.Equal(*tc.fwRevision, *row.GetFieldUint32("fwRevision"))
			} else {
				t.Nil(row.GetFieldUint32("fwRevision"))
			}
			if tc.cfgIdentifier != nil {
				t.Equal(*tc.cfgIdentifier, *row.GetFieldUint32("cfgIdentifier"))
			} else {
				t.Nil(row.GetFieldUint32("cfgIdentifier"))
			}
			if tc.appState != nil {
				t.Equal(*tc.appState, *row.GetFieldUint32("appState"))
			} else {
				t.Nil(row.GetFieldUint32("appState"))
			}
		})
	}
}

func (t *TestSuite) TestDataMessage_BuildConvertAndParse_Types() {
	testCases := []struct {
		name             string
		dataField        func(name string) m2cp.DataField
		dataType         string
		value            interface{}
		expectedDataJson string
		validate         func(t *TestSuite, row m2cp.DataRow)
	}{
		{"bool true", NewDataFieldBool, DataTypeBool, true, `"true"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(true, *row.GetFieldBool("field1"))
			},
		},
		{"bool false", NewDataFieldBool, DataTypeBool, false, `"false"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(false, *row.GetFieldBool("field1"))
			},
		},
		{"int", NewDataFieldInt, DataTypeInt, 42, `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int", NewDataFieldInt, DataTypeInt, Ptr(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int nil", NewDataFieldInt, DataTypeInt, (*int)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldInt("field1"))
			}},
		{"int8", NewDataFieldInt, DataTypeInt, (int8)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int8", NewDataFieldInt, DataTypeInt, Ptr((int8)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int8 nil", NewDataFieldInt, DataTypeInt, (*int8)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldInt("field1"))
			}},
		{"int16", NewDataFieldInt, DataTypeInt, (int16)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int16", NewDataFieldInt, DataTypeInt, Ptr((int16)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int16 nil", NewDataFieldInt, DataTypeInt, (*int16)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldInt("field1"))
			}},
		{"int32", NewDataFieldInt, DataTypeInt, (int32)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int32", NewDataFieldInt, DataTypeInt, Ptr((int32)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int32 nil", NewDataFieldInt, DataTypeInt, (*int32)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldInt("field1"))
			}},
		{"int64", NewDataFieldInt, DataTypeInt, (int64)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int64", NewDataFieldInt, DataTypeInt, Ptr((int64)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*int64 nil", NewDataFieldInt, DataTypeInt, (*int64)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldInt("field1"))
			}},
		{"uint", NewDataFieldInt, DataTypeInt, (uint)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*uint", NewDataFieldInt, DataTypeInt, Ptr((uint)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*uint nil", NewDataFieldInt, DataTypeInt, (*uint)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldInt("field1"))
			}},
		{"uint8", NewDataFieldInt, DataTypeInt, (uint8)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*uint8", NewDataFieldInt, DataTypeInt, Ptr((uint8)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*uint8 nil", NewDataFieldInt, DataTypeInt, (*uint8)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldInt("field1"))
			}},
		{"uint16", NewDataFieldInt, DataTypeInt, (uint16)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*uint16", NewDataFieldInt, DataTypeInt, Ptr((uint16)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(42, *row.GetFieldInt("field1"))
			}},
		{"*uint16 nil", NewDataFieldInt, DataTypeInt, (*uint16)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldInt("field1"))
			}},
		{"uint32", NewDataFieldUint32, DataTypeUint32, (uint32)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal((uint32)(42), *row.GetFieldUint32("field1"))
			}},
		{"*uint32", NewDataFieldUint32, DataTypeUint32, Ptr((uint32)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal((uint32)(42), *row.GetFieldUint32("field1"))
			}},
		{"*uint32 nil", NewDataFieldUint32, DataTypeUint32, (*uint32)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldUint32("field1"))
			}},
		{"uint64", NewDataFieldUint64, DataTypeUint64, (uint64)(42), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal((uint64)(42), *row.GetFieldUint64("field1"))
			}},
		{"*uint64", NewDataFieldUint64, DataTypeUint64, Ptr((uint64)(42)), `"42"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal((uint64)(42), *row.GetFieldUint64("field1"))
			}},
		{"*uint64 nil", NewDataFieldUint64, DataTypeUint64, (*uint64)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldUint64("field1"))
			}},
		{"int array", NewDataFieldIntList, DataTypeIntList, []int{1, 2, 3}, `"1,2,3"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal([]int{1, 2, 3}, row.GetFieldIntArray("field1"))
			}},
		{"int array empty", NewDataFieldIntList, DataTypeIntList, []int{}, `""`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Empty(row.GetFieldIntArray("field1"))
				t.NotNil(row.GetFieldIntArray("field1"))
			}},
		{"int array nil", NewDataFieldIntList, DataTypeIntList, ([]int)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.T().Skip("nil arrays currently not implemented")
				t.Nil(row.GetFieldIntArray("field1"))
			}},
		{"uint32 array", NewDataFieldUint32List, DataTypeUint32List, []uint32{1, 2, 3}, `"1,2,3"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal([]uint32{1, 2, 3}, row.GetFieldUint32Array("field1"))
			}},
		{"uint32 array empty", NewDataFieldUint32List, DataTypeUint32List, []uint32{}, `""`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Empty(row.GetFieldUint32Array("field1"))
				t.NotNil(row.GetFieldUint32Array("field1"))
			}},
		{"uint32 array nil", NewDataFieldUint32List, DataTypeUint32List, ([]uint32)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.T().Skip("nil arrays currently not implemented")
				t.Nil(row.GetFieldUint32Array("field1"))
			}},
		{"uint64 array", NewDataFieldUint64List, DataTypeUint64List, []uint64{1, 2, 3}, `"1,2,3"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal([]uint64{1, 2, 3}, row.GetFieldUint64Array("field1"))
			}},
		{"uint64 array empty", NewDataFieldUint64List, DataTypeUint64List, []uint64{}, `""`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Empty(row.GetFieldUint64Array("field1"))
				t.NotNil(row.GetFieldUint64Array("field1"))
			}},
		{"uint64 array nil", NewDataFieldUint64List, DataTypeUint64List, ([]uint64)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.T().Skip("nil arrays currently not implemented")
				t.Nil(row.GetFieldUint64Array("field1"))
			}},
		{"float32", NewDataFieldDouble, DataTypeDouble, float32(1.5), `"1.5"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(1.5, *row.GetFieldDouble("field1"))
			}},
		{"*float32", NewDataFieldDouble, DataTypeDouble, Ptr(float32(1.5)), `"1.5"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(1.5, *row.GetFieldDouble("field1"))
			}},
		{"*float32 nil", NewDataFieldDouble, DataTypeDouble, (*float32)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldDouble("field1"))
			}},
		{"float32 array", NewDataFieldDoubleList, DataTypeDoubleList, []float32{1.1, 1.2, 1.3}, `"1.1,1.2,1.3"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal([]float64{1.1, 1.2, 1.3}, row.GetFieldDoubleArray("field1"))
			}},
		{"float32 array empty", NewDataFieldDoubleList, DataTypeDoubleList, []float32{}, `""`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Empty(row.GetFieldDoubleArray("field1"))
				t.NotNil(row.GetFieldDoubleArray("field1"))
			}},
		{"float32 array nil", NewDataFieldDoubleList, DataTypeDoubleList, ([]float32)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.T().Skip("nil arrays currently not implemented")
				t.Nil(row.GetFieldDoubleArray("field1"))
			}},
		{"float64", NewDataFieldDouble, DataTypeDouble, float64(1.5), `"1.5"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(1.5, *row.GetFieldDouble("field1"))
			}},
		{"*float64", NewDataFieldDouble, DataTypeDouble, Ptr(float64(1.5)), `"1.5"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(1.5, *row.GetFieldDouble("field1"))
			}},
		{"*float64 nil", NewDataFieldDouble, DataTypeDouble, (*float64)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldDouble("field1"))
			}},
		{"float64 array", NewDataFieldDoubleList, DataTypeDoubleList, []float64{1.1, 1.2, 1.3}, `"1.1,1.2,1.3"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal([]float64{1.1, 1.2, 1.3}, row.GetFieldDoubleArray("field1"))
			}},
		{"float64 array empty", NewDataFieldDoubleList, DataTypeDoubleList, []float64{}, `""`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Empty(row.GetFieldDoubleArray("field1"))
				t.NotNil(row.GetFieldDoubleArray("field1"))
			}},
		{"float64 array nil", NewDataFieldDoubleList, DataTypeDoubleList, ([]float64)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.T().Skip("nil arrays currently not implemented")
				t.Nil(row.GetFieldDoubleArray("field1"))
			}},
		{"string", NewDataFieldString, DataTypeString, "foo", `"foo"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal("foo", *row.GetFieldString("field1"))
			}},
		{"*string", NewDataFieldString, DataTypeString, Ptr("foo"), `"foo"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal("foo", *row.GetFieldString("field1"))
			}},
		{"*string nil", NewDataFieldString, DataTypeString, (*string)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldString("field1"))
			}},
		{"binary", NewDataFieldBinary, DataTypeBinary, []byte{0, 1, 2, 3, 4}, `"AAECAwQ="`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal([]byte{0, 1, 2, 3, 4}, row.GetFieldBinary("field1"))
			}},
		{"binary empty", NewDataFieldBinary, DataTypeBinary, []byte{}, `""`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Empty(row.GetFieldBinary("field1"))
				t.NotNil(row.GetFieldBinary("field1"))
			}},
		{"binary nil", NewDataFieldBinary, DataTypeBinary, ([]byte)(nil), `null`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Nil(row.GetFieldBinary("field1"))
			}},
		{"datetime utc", NewDataFieldDatetime, DataTypeDatetime, time.Unix(1600000000, 1).UTC(), `"2020-09-13T12:26:40.000000001Z"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(time.Unix(1600000000, 1).UTC(), *row.GetFieldDatetime("field1"))
			}},
		{"datetime with local timezone", NewDataFieldDatetime, DataTypeDatetime, time.Unix(1600000000, 1), `"2020-09-13T12:26:40.000000001Z"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(time.Unix(1600000000, 1), *row.GetFieldDatetime("field1"))
			}},
		// time variables and struct fields should be of type [time.Time], not *time.Time; there seems to be no use for nilable datetime -- at this time
		{"datetime zero", NewDataFieldDatetime, DataTypeDatetime, time.Time{}, `"0001-01-01T00:00:00Z"`,
			func(t *TestSuite, row m2cp.DataRow) {
				t.Equal(time.Time{}, *row.GetFieldDatetime("field1"))
				t.True(row.GetFieldDatetime("field1").IsZero())
			}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func() {
			if strings.Contains(tc.name, "array nil") {
				t.T().Skip("nil arrays currently not implemented")
			}
			if "binary nil" == tc.name {
				t.T().Skip("nil binary currently not implemented")
			}
			if "datetime with local timezone" == tc.name {
				t.T().Skip("parsing datetimes with local timezone currently not supported, all datetimes are converted to be in UTC")
			}

			format, err := NewDataFormat("testFormat", tc.dataField("field1"))
			t.NoError(err)

			values := map[string]interface{}{
				"field1": tc.value,
			}

			row := helperBuildConvertAndParseMessage(t, format, "testFormat", values, tc.dataType, tc.expectedDataJson)

			tc.validate(t, row)
		})
	}
}

func helperBuildConvertAndParseMessage(t *TestSuite, format m2cp.DataFormat, formatId string, values map[string]interface{}, dataType string, expectedDataJson string) m2cp.DataRow {
	t.T().Helper()
	rowToSend, err := format.NewRow(values)
	t.Require().NoError(err)
	rowTime := time.Unix(1600000043, 4000)
	rowToSend.SetTime(rowTime)

	addr, err := NewAddressWithSubtopic(t.ctp, "foo.bar.local", "testtopic")
	t.Require().NoError(err)
	msg, err := NewDataMessage(addr, format)
	t.Require().NoError(err)
	msg.GetHeader().SetTimestamp(1680075472.77354)
	msg.GetHeader().SetId("99035d37-0455-439c-b6a7-146565d9951e")
	msg.GetHeader().SetTaskId("eb94490d-306b-4475-9e8b-5b470d669f80")
	t.Equal(formatId, msg.GetFormat().GetFormatId())

	msg.AddRow(rowToSend)
	json, err := ToJson(msg)
	t.JSONEq(fmt.Sprintf(`{
"Header": {
  "Type":"DATA",
  "Topic":"data/testtopic",
  "Scope":"GLOBAL",
  "Timestamp":"1680075472.77354",
  "Id":"99035d37-0455-439c-b6a7-146565d9951e",
  "TaskId":"eb94490d-306b-4475-9e8b-5b470d669f80",
  "Origin":"%s"
},"Body": {
  "FormatId":"%s",
  "Columns":[
  {"Name":"field1","Type":"%s"}],
  "Data":[
  {"T":"1600000043.000004","D":[%s]}
 ]}
}`, addr.GetAddress(), formatId, dataType, expectedDataJson), string(json))
	t.Require().NoError(err)

	parsed, err := DataFromJson(json)
	t.Require().NoError(err)
	t.Equal("testFormat", parsed.GetFormat().GetFormatId())
	rows := parsed.GetRows()
	return rows[0]
}

func Ptr[T any](v T) *T {
	return &v
}

func (t *TestSuite) TestParseAggregatedDataMessage() {
	//const msg1 = `{"Header":{"Type":"DATA","Topic":"data/bar","Timestamp":"1680075472.77354","Id":"dfccbdf5-aa49-4529-af83-ffdda81cdee1","TaskId":"eb94490d-306b-4475-9e8b-5b470d669f80","Origin":"foo"},"Body":{"FormatId":"testFormat","Columns":{"Fields":[{"Name":"good","Type":"bool","Position":0},{"Name":"x","Type":"int","Position":1},{"Name":"xs","Type":"int[]","Position":2},{"Name":"d","Type":"double","Position":3},{"Name":"ds","Type":"double[]","Position":4},{"Name":"s","Type":"string","Position":5},{"Name":"b","Type":"binary","Position":6},{"Name":"t","Type":"datetime","Position":7}]},"Data":[{"T":"1600000042.000003","D":["true","42","1,2,3","1.5","1.1,1.2,1.3","foo","AAECAwQ=","2020-09-13T12:26:40.000000001Z"]},{"T":"1600000042.000003","D":["true","42","1,2,3","1.5","1.1,1.2,1.3","foo","AAECAwQ=","2020-09-13T12:26:40.000000001Z"]},{"T":"1600000043.000004","D":["false","1","1,2,3","1.5","1.21,441.2,-551.3","hello","KhEMInw=","2020-09-13T12:26:40.000002Z"]}]},"Assertion":{}}`
	//parsed1, err := DataFromJson([]byte(msg1))
	//s.NoError(err)
	//s.Equal("testFormat", parsed1.GetFormat().GetFormatId())
	const msg2 = `{"Header":{"Type":"DATA","Topic":"data/simsensor-7527F8A4/","Timestamp":"1686660497.4338222","Id":"582b8980-5daf-4815-8f52-08837522e95c","Origin":"simsensor-7527F8A4.development-654331.9c12c037-c379-43c4-bad9-3dd992cd9ade"},"Body":{"FormatId":"7f7778a3-8509-4124-9433-3ba93e5f1c43","Columns":[{"Name":"counter","Type":"int"},{"Name":"temperature","Type":"double"}],"Data":[{"T":"1686660492.3643131","D":["0","18.511042502439338"]},{"T":"1686660497.3667634","D":["1","18.05715491178404"]}]}}`
	parsed2, err := DataFromJson([]byte(msg2))
	t.NoError(err)
	t.Equal("7f7778a3-8509-4124-9433-3ba93e5f1c43", parsed2.GetFormat().GetFormatId())
}
