package pyadapter

import (
	"m2cp"
	"m2cp/messages"
)

type DataMessagePyAdapter struct {
	payload m2cp.DataMessage
}

func (dp DataMessagePyAdapter) GetPayload() m2cp.DataMessage {
	return dp.payload
}

func (dp DataMessagePyAdapter) GetDataFormatPyAdapter() DataFormatPyAdapter {
	return DataFormatPyAdapter{dp.GetPayload().GetFormat()}
}

type DataFormatPyAdapter struct {
	payload m2cp.DataFormat
}

func (df DataFormatPyAdapter) GetPayload() m2cp.DataFormat {
	return df.payload
}

func NewDataFormatPyAdapter(formatId string, fields ...m2cp.DataField) (DataFormatPyAdapter, error) {
	df, err := messages.NewDataFormat(formatId, fields...)
	return DataFormatPyAdapter{payload: df}, err
}

func (df DataFormatPyAdapter) GetFieldsPyAdapter() []DataFieldPyAdapter {
	res := []DataFieldPyAdapter{}
	for _, f := range df.GetPayload().GetFields() {
		res = append(res, DataFieldPyAdapter{payload: f})
	}
	return res
}

type DataFieldPyAdapter struct {
	payload m2cp.DataField
}

func (df DataFieldPyAdapter) GetPayload() m2cp.DataField {
	return df.payload
}

func SubscribeData(connection m2cp.NetworkConnection, topics []string, callback func(adapter *DataMessagePyAdapter)) error {
	handler := func(message m2cp.DataMessage) {
		adapter := DataMessagePyAdapter{message}
		callback(&adapter)
	}

	return connection.SubscribeData(topics, handler)
}

func SubscribeDataWithOptions(connection m2cp.NetworkConnection, topics []string, callback func(adapter *DataMessagePyAdapter), options m2cp.SubscriptionOptions) error {
	handler := func(message m2cp.DataMessage) {
		adapter := DataMessagePyAdapter{message}
		callback(&adapter)
	}

	return connection.SubscribeDataWithOptions(topics, handler, options)
}
