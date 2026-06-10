package messages

import (
	"encoding/base64"
	"errors"
	"fmt"
	"m2cp"
	"m2cp/tools"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	DataTypeString     = "string"
	DataTypeBool       = "bool"
	DataTypeInt        = "int"
	DataTypeIntList    = "int[]"
	DataTypeUint32     = "uint32"
	DataTypeUint32List = "uint32[]"
	DataTypeUint64     = "uint64"
	DataTypeUint64List = "uint64[]"
	DataTypeDouble     = "double"
	DataTypeDoubleList = "double[]"
	DataTypeBinary     = "binary"
	DataTypeDatetime   = "datetime"
)

var validDataTypes = []string{
	DataTypeString,
	DataTypeBool,
	DataTypeInt, DataTypeIntList,
	DataTypeUint32, DataTypeUint32List,
	DataTypeUint64, DataTypeUint64List,
	DataTypeDouble, DataTypeDoubleList,
	DataTypeBinary,
	DataTypeDatetime,
}

var nextID int64 = 0

type dataFormat struct {
	fields            []dataField
	formatId          string
	internalID        int64
	fieldsByName      map[string]*dataField
	fieldsByPos       map[int]*dataField
	dataQueueMaxAge   time.Duration
	dataQueueMaxCount int
	messageScope      m2cp.MessageScope
}

func (df *dataFormat) NewRow(items map[string]interface{}) (m2cp.DataRow, error) {
	itemsList := make([]string, len(df.fields))
	for i := 0; i < len(df.fields); i++ {
		field := df.fields[i]
		fieldName := field.GetName()
		if fieldValue, ok := items[fieldName]; !ok {
			return nil, errors.New(fmt.Sprintf("field %s not found in items", fieldName))
		} else {
			fieldValueEncoded, detectedType := stringify(fieldValue)
			if fieldValueEncoded == nil && detectedType == nil {
				return nil, errors.New(fmt.Sprintf("field %s has invalid type %v - failed encoding '%v' to string", fieldName, detectedType, fieldValue))
			}
			if fieldValueEncoded == nil {
				// allow optional values; if value is nil, we treat it as null and set the field value to empty string (so that it can be deserialized back to null on the receiving end)
				// all getters will return nil for this value, regardless of the field type,
				// so it can be used for any field type except string, for which it will be treated as the literal string "null"
				itemsList[i] = "null"
				continue
			}
			if *detectedType != field.GetType() {
				return nil, errors.New(fmt.Sprintf("field %s has invalid type - expected %s, got %s", fieldName, field.GetType(), *detectedType))
			}
			itemsList[i] = *fieldValueEncoded
		}
	}
	return NewDataRow(df, time.Now(), itemsList), nil
}

// NewRowFromPreStringifiedFieldValues WARNING: power user feature, used by python SDK wrapper ... skips the type checking for the values
func (df *dataFormat) NewRowFromPreStringifiedFieldValues(items map[string]string) (m2cp.DataRow, error) {
	itemsList := make([]string, len(df.fields))
	for i := 0; i < len(df.fields); i++ {
		field := df.fields[i]
		fieldName := field.GetName()
		if fieldValue, ok := items[fieldName]; !ok {
			return nil, errors.New(fmt.Sprintf("field %s not found in items", fieldName))
		} else {

			itemsList[i] = fieldValue

		}
	}
	return NewDataRow(df, time.Now(), itemsList), nil
}

func (df *dataFormat) SetDataQueueMaxAge(duration time.Duration) {
	df.dataQueueMaxAge = duration
}

func (df *dataFormat) GetDataQueueMaxAge() time.Duration {
	return df.dataQueueMaxAge
}

func (df *dataFormat) SetDataQueueMaxCount(length int) {
	if length < 0 {
		length = 0
	}
	df.dataQueueMaxCount = length
}

func (df *dataFormat) GetDataQueueMaxCount() int {
	return df.dataQueueMaxCount
}

func (df *dataFormat) GetInternalId() int64 {
	return df.internalID
}

func (df *dataFormat) GetFormatId() string {
	return df.formatId
}

func (df *dataFormat) GetFieldPos(field string) int {
	if f, ok := df.fieldsByName[field]; ok {
		return f.GetPos()
	} else {
		return -1
	}
}

func (df *dataFormat) GetFields() []m2cp.DataField {
	fields := make([]m2cp.DataField, len(df.fields))
	for i := 0; i < len(df.fields); i++ {
		field := df.fields[i]
		fields[i] = &field
	}
	return fields
}

func (df *dataFormat) getColumns() []dataField {
	return df.fields
}

type dataField struct {
	Name     string
	Type     string
	position int
}

func NewDataFieldBool(name string) m2cp.DataField {
	return newDataField(name, DataTypeBool)
}

func NewDataFieldInt(name string) m2cp.DataField {
	return newDataField(name, DataTypeInt)
}

func NewDataFieldIntList(name string) m2cp.DataField {
	return newDataField(name, DataTypeIntList)
}

func NewDataFieldUint32(name string) m2cp.DataField {
	return newDataField(name, DataTypeUint32)
}

func NewDataFieldUint32List(name string) m2cp.DataField {
	return newDataField(name, DataTypeUint32List)
}

func NewDataFieldUint64(name string) m2cp.DataField {
	return newDataField(name, DataTypeUint64)
}

func NewDataFieldUint64List(name string) m2cp.DataField {
	return newDataField(name, DataTypeUint64List)
}

func NewDataFieldDouble(name string) m2cp.DataField {
	return newDataField(name, DataTypeDouble)
}

func NewDataFieldDoubleList(name string) m2cp.DataField {
	return newDataField(name, DataTypeDoubleList)
}

func NewDataFieldString(name string) m2cp.DataField {
	return newDataField(name, DataTypeString)
}

func NewDataFieldBinary(name string) m2cp.DataField {
	return newDataField(name, DataTypeBinary)
}

func NewDataFieldDatetime(name string) m2cp.DataField {
	return newDataField(name, DataTypeDatetime)
}

func newDataField(name, fieldType string) m2cp.DataField {
	f := dataField{
		Name:     name,
		Type:     fieldType,
		position: -1,
	}
	return &f
}

func (df *dataField) GetName() string {
	return df.Name
}

func (df *dataField) GetType() string {
	return df.Type
}

func (df *dataField) GetPos() int {
	return df.position
}

func (df *dataField) SetPos(pos int) {
	df.position = pos
}

func NewDataFormat(formatId string, fields ...m2cp.DataField) (m2cp.DataFormat, error) {
	fs := make([]dataField, len(fields))
	for i := 0; i < len(fields); i++ {
		fs[i] = *fields[i].(*dataField)
	}
	df := dataFormat{}
	df.InitAfterParsing(formatId, fs)
	return &df, nil
}

func (df *dataFormat) CountBytes() int {
	// TODO: implement
	return 0
}

func (df *dataFormat) GetScope() m2cp.MessageScope {
	return df.messageScope
}

func (df *dataFormat) SetScope(scope m2cp.MessageScope) {
	df.messageScope = scope

}

func (df *dataFormat) InitAfterParsing(formatId string, fields []dataField) {
	df.internalID = atomic.AddInt64(&nextID, 1)
	df.formatId = formatId
	if df.messageScope == m2cp.MessageScopeUndefined {
		df.messageScope = m2cp.MessageScopeGlobal
	}
	df.fields = fields
	df.fieldsByName = make(map[string]*dataField)
	df.fieldsByPos = make(map[int]*dataField)
	for i := range df.fields {
		field := &df.fields[i]
		field.SetPos(i)
		df.fieldsByName[field.GetName()] = field
		df.fieldsByPos[field.GetPos()] = field
	}
	// default rows-aggregation values
	df.dataQueueMaxAge = 5000 * time.Millisecond
	df.dataQueueMaxCount = 100
}

func isValidDataType(dataType string) bool {
	for _, t := range validDataTypes {
		if t == dataType {
			return true
		}
	}
	return false
}

func stringify(value interface{}) (*string, *string) {
	var s string
	var t string
	switch v := value.(type) {
	case string:
		s = v
		t = DataTypeString
	case bool:
		s = strconv.FormatBool(v)
		t = DataTypeBool
	case int:
		s = strconv.Itoa(v)
		t = DataTypeInt
	case int8:
		s = strconv.Itoa(int(v))
		t = DataTypeInt
	case int16:
		s = strconv.Itoa(int(v))
		t = DataTypeInt
	case int32:
		s = strconv.Itoa(int(v))
		t = DataTypeInt
	case int64:
		s = strconv.FormatInt(v, 10)
		t = DataTypeInt
	case *int:
		t = DataTypeInt
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case *int8:
		t = DataTypeInt
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case *int16:
		t = DataTypeInt
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case *int32:
		t = DataTypeInt
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case *int64:
		t = DataTypeInt
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case uint:
		s = strconv.FormatUint(uint64(v), 10)
		t = DataTypeInt
	case uint8:
		s = strconv.FormatUint(uint64(v), 10)
		t = DataTypeInt
	case uint16:
		s = strconv.FormatUint(uint64(v), 10)
		t = DataTypeInt
	case uint32:
		s = strconv.FormatUint(uint64(v), 10)
		t = DataTypeUint32
	case uint64:
		s = strconv.FormatUint(v, 10)
		t = DataTypeUint64
	case *uint:
		t = DataTypeInt
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case *uint8:
		t = DataTypeInt
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case *uint16:
		t = DataTypeInt
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case *uint32:
		t = DataTypeUint32
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case *uint64:
		t = DataTypeUint64
		if v == nil {
			return nil, &t
		}
		return stringify(*v)
	case []int:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeIntList
	case []int8:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeIntList
	case []int16:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeIntList
	case []int32:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeIntList
	case []int64:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeIntList
	case []uint:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeIntList
	case []uint16:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeIntList
	case []uint32:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeUint32List
	case []uint64:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeUint64List
	case float32:
		s = strconv.FormatFloat(float64(v), 'g', -1, 64)
		t = DataTypeDouble
	case float64:
		s = strconv.FormatFloat(v, 'g', -1, 64)
		t = DataTypeDouble
	case []float32:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeDoubleList
	case []float64:
		s = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v)), ","), "[]")
		t = DataTypeDoubleList
	case []byte:
		s = base64.StdEncoding.EncodeToString(v)
		t = DataTypeBinary
	case time.Time:
		// convert to (ISO 8601 in UTC) format
		s = tools.TimeToISO8601UTC(v)
		t = DataTypeDatetime
	default:
		return nil, nil
	}
	return &s, &t
}

func GetInternalFormatId(f m2cp.DataFormat) int64 {
	return f.(*dataFormat).internalID
}
