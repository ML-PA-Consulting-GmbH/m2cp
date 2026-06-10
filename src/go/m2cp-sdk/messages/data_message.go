package messages

import (
	"m2cp"
)

type dataMessage struct {
	Header messageHeader `json:"Header"`
	Body   dataBody      `json:"Body"`
	Trace  []trace       `json:"Trace,omitempty"`
}

type dataBody struct {
	FormatId string
	Columns  []dataField `json:"Columns"`
	Data     []dataRow
	format   *dataFormat
	bytes    int
}

func (b *dataBody) initAfterParsing() error {
	cols := make([]m2cp.DataField, len(b.Columns))
	for i := 0; i < len(b.Columns); i++ {
		cols[i] = &b.Columns[i]
	}
	f, err := NewDataFormat(b.FormatId, cols...)
	if err != nil {
		return err
	}
	b.format = f.(*dataFormat)
	b.bytes = 42 + len(b.FormatId) + b.format.CountBytes()
	for k := range b.Data {
		var row = &b.Data[k]
		b.bytes += row.CountBytes()
		row.format = b.format
	}
	b.format.formatId = b.FormatId
	return nil
}

func NewDataMessage(from m2cp.Address, format m2cp.DataFormat) (m2cp.DataMessage, error) {
	m := &dataMessage{
		Header: messageHeader{
			Type:   "DATA",
			Topic:  "data/" + from.GetSubtopic(),
			Scope:  format.GetScope().String(),
			Origin: from.GetAddress(),
		},
		Body: dataBody{
			FormatId: format.GetFormatId(),
			format:   format.(*dataFormat),
			Columns:  format.(*dataFormat).getColumns(),
			Data:     []dataRow{},
			bytes:    0,
		},
		Trace: []trace{},
	}
	m.Header.SetTimestampNow()
	m.Header.SetIdRandom()
	m.Header.SetTaskIdRandom()
	return m, nil
}

func (o *dataMessage) GetHeader() m2cp.MessageHeader {
	return &o.Header
}

func (o *dataMessage) GetTraces() []m2cp.MessageTrace {
	var traces = make([]m2cp.MessageTrace, len(o.Trace))
	for i := range o.Trace {
		traces[i] = &o.Trace[i]
	}
	return traces
}

func (o *dataMessage) AddTrace(relay string) {
	o.Trace = append(o.Trace, newTrace(relay))
}

func (o *dataMessage) GetFormat() m2cp.DataFormat {
	return o.Body.format
}

func (o *dataMessage) CountRows() int {
	return len(o.Body.Data)
}

func (o *dataMessage) GetRows() []m2cp.DataRow {
	rows := make([]m2cp.DataRow, len(o.Body.Data))
	for k := range o.Body.Data {
		var row *dataRow = &o.Body.Data[k]
		rows[k] = row
	}
	return rows
}

func (o *dataMessage) AddRow(row m2cp.DataRow) {
	o.Body.Data = append(o.Body.Data, *row.(*dataRow))
	//m.Body.bytes += row.GetSize()
}

func (o *dataMessage) AddRows(rows []m2cp.DataRow) {
	rs := make([]dataRow, len(rows))
	for k := range rows {
		rs[k] = *rows[k].(*dataRow)
	}
	o.Body.Data = append(o.Body.Data, rs...)
	//for _, row := range rows {
	//	m.Body.bytes += row.GetSize()
	//}
}

func (o *dataMessage) initAfterParsing() error {
	err := o.Body.initAfterParsing()
	o.Body.format.SetScope(o.Header.GetScope())
	return err
}
