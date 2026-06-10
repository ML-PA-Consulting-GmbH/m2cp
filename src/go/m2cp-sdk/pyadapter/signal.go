package pyadapter

import "m2cp"

// SubscribeSignals(topics []string, handler SignalHandler) error
//	SubscribeSignalsWithOptions(topics []string, handler SignalHandler, options SubscriptionOptions) error
//	SubscribeData(topics []string, handler DataHandler) error
//	SubscribeDataWithOptions(topics []string, handler DataHandler, options SubscriptionOptions) error

type SignalMessagePyAdapter struct {
	payload m2cp.SignalMessage
}

func (sp SignalMessagePyAdapter) GetHeader() m2cp.MessageHeader {
	return sp.payload.GetHeader()
}
func (sp SignalMessagePyAdapter) GetTraces() []m2cp.MessageTrace {
	return sp.payload.GetTraces()
}
func (sp SignalMessagePyAdapter) AddTrace(relay string) {
	sp.payload.AddTrace(relay)
}
func (sp SignalMessagePyAdapter) GetType() string {
	return sp.payload.GetType()
}
func (sp SignalMessagePyAdapter) GetName() string {
	return sp.payload.GetName()
}
func (sp SignalMessagePyAdapter) GetContent() string {
	return sp.payload.GetContent()
}

func SubscribeSignals(connection m2cp.NetworkConnection, topics []string, callback func(adapter *SignalMessagePyAdapter)) error {
	handler := func(message m2cp.SignalMessage) {
		adapter := SignalMessagePyAdapter{message}
		callback(&adapter)
	}

	return connection.SubscribeSignals(topics, handler)
}

func SubscribeSignalsWithOptions(connection m2cp.NetworkConnection, topics []string, callback func(adapter *SignalMessagePyAdapter), options m2cp.SubscriptionOptions) error {
	handler := func(message m2cp.SignalMessage) {
		adapter := SignalMessagePyAdapter{message}
		callback(&adapter)
	}

	return connection.SubscribeSignalsWithOptions(topics, handler, options)
}
