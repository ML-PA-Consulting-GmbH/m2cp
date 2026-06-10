package amqp

import (
	"fmt"
	"m2cp"
	"m2cp/messages"
	"time"
)

type SendJob struct {
	Exchange        string
	RoutingKey      string
	Topic           string
	Message         []byte
	MessageType     m2cp.MessageType
	MessageID       string
	Origin          string
	CorrelationId   string
	DeliveryTag     uint64
	DeliveryAttempt time.Duration
}

func newSendJobs(message m2cp.Message, serializer m2cp.MessageSerializer) ([]*SendJob, error) {
	messageJson, err := messages.ToBinary(message, serializer)
	if err != nil {
		return nil, err
	}
	h := message.GetHeader()
	if h == nil {
		return nil, fmt.Errorf("message header is nil")
	}
	topic := h.GetTopic()
	exchange, _ := topic2Exchange(topic)
	routingKey := topic2RoutingKey(topic)
	via := message.GetHeader().GetVia()
	if via != "" {
		routingKey = via
	}

	sj := &SendJob{
		MessageType: message.GetHeader().GetType(),
		Exchange:    exchange,
		RoutingKey:  routingKey,
		Topic:       topic,
		Message:     messageJson,
		MessageID:   h.GetId(),
		Origin:      h.GetName(),
	}

	switch sj.MessageType {
	case m2cp.MessageTypeData:
		return []*SendJob{sj}, nil
	case m2cp.MessageTypeSignal:
		return []*SendJob{sj}, nil
	case m2cp.MessageTypeCommand:
		sj.CorrelationId = message.GetHeader().GetId()
		sjLog := *sj
		sjLog.Exchange = ExchangeNameLog
		sjLog.Topic = "log/" + topic
		sjLog.RoutingKey = topic2RoutingKey(sjLog.Topic)
		return []*SendJob{sj, &sjLog}, nil
	case m2cp.MessageTypeResponse:
		sj.CorrelationId = message.(m2cp.ResponseMessage).GetCommandId()
		sjLog := *sj
		sjLog.Exchange = ExchangeNameLog
		sjLog.Topic = "log/" + topic
		sjLog.RoutingKey = topic2RoutingKey(sjLog.Topic)
		return []*SendJob{sj, &sjLog}, nil
	}
	return nil, fmt.Errorf("unknown message type: %s", sj.MessageType)
}

func (s SendJob) String() string {
	return fmt.Sprintf("SendJob{Exchange: %s, RoutingKey: %s, Topic: %s, MessageType: %s, MessageID: %s, Origin: %s, CorrelationId: %s}", s.Exchange, s.RoutingKey, s.Topic, s.MessageType, s.MessageID, s.Origin, s.CorrelationId)
}
