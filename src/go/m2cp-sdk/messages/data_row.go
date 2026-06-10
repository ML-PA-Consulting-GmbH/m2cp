package messages

import (
	"encoding/base64"
	"m2cp"
	"m2cp/tools"
	"strconv"
	"strings"
	"time"
)

type dataRow struct {
	T           string
	D           []string
	format      *dataFormat
	timeCreated int64
	bytes       int
}

func NewDataRow(format *dataFormat, timestamp time.Time, items []string) m2cp.DataRow {
	timeCreated := time.Now().UnixNano() / int64(time.Millisecond)
	timeString := tools.TimeToDec(timestamp)
	bytes := 15 + len(timeString) + stringArrayBytes(items) + 7*len(items)

	return &dataRow{
		T:           timeString,
		D:           items,
		format:      format,
		timeCreated: timeCreated,
		bytes:       bytes,
	}
}

func (dr *dataRow) GetFormat() m2cp.DataFormat {
	return dr.format
}

func (dr *dataRow) GetTimeString() string {
	return dr.T
}

func (dr *dataRow) GetFieldsRaw() map[string]string {
	fields := dr.format.GetFields()
	fieldsRaw := make(map[string]string)
	for _, field := range fields {
		fieldsRaw[field.GetName()] = dr.D[field.GetPos()]
	}
	return fieldsRaw
}

func (dr *dataRow) GetTime() time.Time {
	return tools.TimeDecToTime(dr.T)
}

func (dr *dataRow) SetTime(timestamp time.Time) {
	dr.T = tools.TimeToDec(timestamp)
}

func (dr *dataRow) GetFieldInt(field string) *int {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	i, err := strconv.Atoi(*v)
	if err != nil {
		return nil
	}
	return &i
}

func (dr *dataRow) GetFieldIntArray(field string) []int {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	strs := strings.Split(*v, ",")
	ints := make([]int, len(strs))
	for i, s := range strs {
		ints[i], _ = strconv.Atoi(s)
	}
	return ints
}

func (dr *dataRow) GetFieldLong(field string) *int64 {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	l, _ := strconv.ParseInt(*v, 10, 64)
	return &l
}

func (dr *dataRow) GetFieldUint32(field string) *uint32 {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	u, err := strconv.ParseUint(*v, 10, 32)
	if err != nil {
		return nil
	}
	u32 := uint32(u)
	return &u32
}

func (dr *dataRow) GetFieldUint32Array(field string) []uint32 {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	strs := strings.Split(*v, ",")
	uint32s := make([]uint32, len(strs))
	for i, s := range strs {
		u32, _ := strconv.ParseUint(s, 10, 32)
		uint32s[i] = uint32(u32)
	}
	return uint32s
}

func (dr *dataRow) GetFieldUint64(field string) *uint64 {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	u, err := strconv.ParseUint(*v, 10, 64)
	if err != nil {
		return nil
	}
	return &u
}

func (dr *dataRow) GetFieldUint64Array(field string) []uint64 {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	strs := strings.Split(*v, ",")
	uint64s := make([]uint64, len(strs))
	for i, s := range strs {
		uint64s[i], _ = strconv.ParseUint(s, 10, 64)
	}
	return uint64s
}

func (dr *dataRow) GetFieldDouble(field string) *float64 {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	f, _ := strconv.ParseFloat(*v, 64)
	return &f
}

func (dr *dataRow) GetFieldDoubleArray(field string) []float64 {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	strs := strings.Split(*v, ",")
	floats := make([]float64, len(strs))
	for i, s := range strs {
		floats[i], _ = strconv.ParseFloat(s, 64)
	}
	return floats
}

func (dr *dataRow) GetFieldBool(field string) *bool {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	b, _ := strconv.ParseBool(*v)
	return &b
}

func (dr *dataRow) GetFieldString(field string) *string {
	i := dr.format.GetFieldPos(field)
	if i >= 0 && i < len(dr.D) {
		res := dr.D[i]
		return &res
	} else {
		return nil
	}
}

func (dr *dataRow) GetFieldBinary(field string) []byte {
	v := dr.GetFieldString(field)
	if v == nil {
		return nil
	}
	bytes, _ := base64.StdEncoding.DecodeString(*v)
	return bytes
}

func (dr *dataRow) GetFieldDatetime(s string) *time.Time {
	v := dr.GetFieldString(s)
	if v == nil {
		return nil
	}
	t, err := tools.TimeParse(*v)
	if err != nil {
		return nil
	}
	return &t
}

func (dr *dataRow) IsFieldNull(field string) bool {
	return dr.GetFieldString(field) == nil
}

func (dr *dataRow) GetAgeMilliseconds() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) - dr.timeCreated
}

func (dr *dataRow) CountBytes() int {
	return dr.bytes
}

func (dr *dataRow) getDataRow() *dataRow {
	return dr
}

func stringArrayBytes(arr []string) int {
	bytes := 0
	for _, s := range arr {
		bytes += len(s)
	}
	return bytes
}
