package messages

import (
	"encoding/base64"
	"fmt"
	"m2cp"
	"m2cp/tools"
	"math/rand"
	"strconv"
	"strings"
	"time"

	fuzz "github.com/google/gofuzz"
)

func (t *TestSuite) TestDataFormat() {
	fields := []m2cp.DataField{
		NewDataFieldBool("testBool"),
	}
	dfA, err := NewDataFormat("testFormat", fields...)
	t.NoError(err)
	dfB, err := NewDataFormat("testFormat", fields...)
	t.NoError(err)
	idA := dfA.(*dataFormat).GetInternalId()
	idB := dfB.(*dataFormat).GetInternalId()
	t.Equal(idA+1, idB)

	fieldsGet := dfA.GetFields()
	for i, field := range fieldsGet {
		t.Equal(field.GetName(), fields[i].GetName())
		t.Equal(field.GetType(), fields[i].GetType())
	}
}

func (t *TestSuite) TestDataFormatScope() {
	df, err := NewDataFormat("testFormat")
	t.NoError(err)
	t.Equal(m2cp.MessageScopeGlobal, df.GetScope())
	df.SetScope(m2cp.MessageScopeDevice)
	t.Equal(m2cp.MessageScopeDevice, df.GetScope())
}

func (t *TestSuite) TestNewDataField() {
	field := NewDataFieldBool("testBool")
	t.NotNil(field)
	t.Equal("testBool", field.GetName())
	t.Equal(DataTypeBool, field.GetType())
}

func (t *TestSuite) TestNewDataFormat() {
	fields := []m2cp.DataField{
		NewDataFieldBool("testBool"),
		NewDataFieldInt("testInt"),
		NewDataFieldString("testString"),
	}

	df, err := NewDataFormat("testFormat", fields...)
	t.NoError(err)
	t.NotNil(df)
	t.Equal("testFormat", df.GetFormatId())
}

func (t *TestSuite) TestNewDataRow() {
	fields := []m2cp.DataField{
		NewDataFieldBool("testBool"),
		NewDataFieldInt("testInt"),
		NewDataFieldString("testString"),
	}

	df, _ := NewDataFormat("testFormat", fields...)

	items := map[string]interface{}{
		"testBool":   true,
		"testInt":    42,
		"testString": "hello",
	}

	dr, err := df.NewRow(items)
	t.NoError(err)
	t.NotNil(dr)
}

func (t *TestSuite) TestValidDataType() {
	validType := DataTypeBool
	t.True(isValidDataType(validType))
	invalidType := "invalidType"
	t.False(isValidDataType(invalidType))
}

func (t *TestSuite) TestStringify() {
	value := "test"
	str, dataType := stringify(value)
	t.NotNil(str)
	t.NotNil(dataType)
	t.Equal("test", *str)
	t.Equal(DataTypeString, *dataType)

	valueTime := time.Now()
	str, dataType = stringify(valueTime)
	t.NotNil(str)
	t.NotNil(dataType)
	t.Equal(tools.TimeToISO8601UTC(valueTime), *str)
	t.Equal(DataTypeDatetime, *dataType)
}

// TestStringify_NilValues tests that stringify allows optional fields to be nil.
func (t *TestSuite) TestStringify_NilValues() {
	t.Run("nil", func() {
		nilStr, nilDataType := stringify(nil)
		t.Nil(nilStr)
		t.Nil(nilDataType)
	})

	testCases := []struct {
		name             string
		value            interface{}
		expectedDataType string
	}{
		//{value: (*string)(nil), expectedDataType: DataTypeString},
		//{value: (*bool)(nil), expectedDataType: DataTypeBool},
		{name: "int", value: (*int)(nil), expectedDataType: DataTypeInt},
		{name: "int8", value: (*int8)(nil), expectedDataType: DataTypeInt},
		{name: "int16", value: (*int16)(nil), expectedDataType: DataTypeInt},
		{name: "int32", value: (*int32)(nil), expectedDataType: DataTypeInt},
		{name: "int64", value: (*int64)(nil), expectedDataType: DataTypeInt},
		{name: "uint", value: (*uint)(nil), expectedDataType: DataTypeInt},
		{name: "uint8", value: (*uint8)(nil), expectedDataType: DataTypeInt},
		{name: "uint16", value: (*uint16)(nil), expectedDataType: DataTypeInt},
		{name: "uint32", value: (*uint32)(nil), expectedDataType: DataTypeUint32},
		{name: "uint64", value: (*uint64)(nil), expectedDataType: DataTypeUint64},
		//{value: (*float64)(nil), expectedDataType: DataTypeDouble},
		//{value: (*time.Time)(nil), expectedDataType: DataTypeDatetime},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func() {
			str, dataType := stringify(tc.value)
			t.Nil(str)
			t.NotNil(dataType)
			t.Equal(tc.expectedDataType, *dataType)
		})
	}
}

func (t *TestSuite) TestStringifyFuzzing() {
	f := fuzz.New()

	// Run tests for different data types
	for i := 0; i < 10000; i++ {
		var v interface{}
		var tp string
		var expected string
		randCase := rand.Intn(28)
		switch dataType := randCase; dataType {
		case 0:
			var value string
			f.Fuzz(&value)
			v = value
			tp = DataTypeString
			expected = value
		case 1:
			var value bool
			f.Fuzz(&value)
			v = value
			tp = DataTypeBool
			expected = strconv.FormatBool(value)
		case 2:
			var value int
			f.Fuzz(&value)
			v = value
			tp = DataTypeInt
			expected = strconv.Itoa(value)
		case 3:
			var value int8
			f.Fuzz(&value)
			v = value
			tp = DataTypeInt
			expected = strconv.Itoa(int(value))
		case 4:
			var value int16
			f.Fuzz(&value)
			v = value
			tp = DataTypeInt
			expected = strconv.Itoa(int(value))
		case 5:
			var value int32
			f.Fuzz(&value)
			v = value
			tp = DataTypeInt
			expected = strconv.Itoa(int(value))
		case 6:
			var value int64
			f.Fuzz(&value)
			v = value
			tp = DataTypeInt
			expected = strconv.FormatInt(value, 10)
		case 7:
			var value uint
			f.Fuzz(&value)
			v = value
			tp = DataTypeInt
			expected = strconv.FormatUint(uint64(value), 10)
		case 8:
			var value uint8
			f.Fuzz(&value)
			v = value
			tp = DataTypeInt
			expected = strconv.FormatUint(uint64(value), 10)
		case 9:
			var value uint16
			f.Fuzz(&value)
			v = value
			tp = DataTypeInt
			expected = strconv.FormatUint(uint64(value), 10)
		case 10:
			var value uint32
			f.Fuzz(&value)
			v = value
			tp = DataTypeUint32
			expected = strconv.FormatUint(uint64(value), 10)
		case 11:
			var value uint64
			f.Fuzz(&value)
			v = value
			tp = DataTypeUint64
			expected = strconv.FormatUint(value, 10)
		case 12:
			var value []int
			f.Fuzz(&value)
			v = value
			tp = DataTypeIntList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 13:
			var value []int8
			f.Fuzz(&value)
			v = value
			tp = DataTypeIntList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 14:
			var value []int16
			f.Fuzz(&value)
			v = value
			tp = DataTypeIntList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 15:
			var value []int32
			f.Fuzz(&value)
			v = value
			tp = DataTypeIntList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 16:
			var value []int64
			f.Fuzz(&value)
			v = value
			tp = DataTypeIntList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 17:
			var value []uint
			f.Fuzz(&value)
			v = value
			tp = DataTypeIntList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 18:
			// uint8 is identical to []byte - so we skip this case
			continue
		case 19:
			var value []uint16
			f.Fuzz(&value)
			v = value
			tp = DataTypeIntList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 20:
			var value []uint32
			f.Fuzz(&value)
			v = value
			tp = DataTypeUint32List
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 21:
			var value []uint64
			f.Fuzz(&value)
			v = value
			tp = DataTypeUint64List
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 22:
			var value float32
			f.Fuzz(&value)
			v = value
			tp = DataTypeDouble
			expected = strconv.FormatFloat(float64(value), 'g', -1, 64)
		case 23:
			var value float64
			f.Fuzz(&value)
			v = value
			tp = DataTypeDouble
			expected = strconv.FormatFloat(value, 'g', -1, 64)
		case 24:
			var value []float32
			f.Fuzz(&value)
			v = value
			tp = DataTypeDoubleList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 25:
			var value []float64
			f.Fuzz(&value)
			v = value
			tp = DataTypeDoubleList
			expected = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
		case 26:
			var value []byte
			f.Fuzz(&value)
			v = value
			tp = DataTypeBinary
			expected = base64.StdEncoding.EncodeToString(value)
		case 27:
			var value time.Time
			f.Fuzz(&value)
			v = value
			tp = DataTypeDatetime
			expected = tools.TimeToISO8601UTC(value)
		default:
			panic("unknown type")
		}
		str, typ := stringify(v)
		t.NotNil(str)
		t.NotNil(typ)
		t.Equal(expected, *str, "case: %d", randCase)
		t.Equal(tp, *typ, "type match failed for case %d, %v != %v", randCase, tp, *typ)
	}
}
