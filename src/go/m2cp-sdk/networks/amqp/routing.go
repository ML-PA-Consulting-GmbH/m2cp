package amqp

import (
	"errors"
	"fmt"
	"m2cp"
	"strings"
)

const (
	ExchangeNameData    = "M2CPSDK-Data"
	ExchangeTypeData    = "topic"
	ExchangeNameSignals = "M2CPSDK-Signals"
	ExchangeTypeSignals = "topic"
	ExchangeNameRpc     = "" // "M2CPSDK-CommandsAndResponses"
	ExchangeTypeRpc     = "direct"
	ExchangeNameLog     = "M2CPSDK-Log"
	ExchangeTypeLog     = "fanout"
	ExchangeNameTest    = "M2CPSDK-Test"
	ExchangeTypeTest    = "fanout"
)

func topic2RoutingKey(topic string) string {
	// Note: older SDK versions replaced * -> #, disabling the use of * as an additional filter option
	//topic = strings.Replace(topic, "*", "#", -1)
	for len(topic) > 0 && topic[len(topic)-1] == '/' {
		topic = topic[:len(topic)-1]
	}
	for len(topic) > 0 && topic[0] == '/' {
		topic = topic[1:]
	}
	topic = strings.Replace(topic, "/", ".", -1)
	if !strings.HasPrefix(topic, "command.") && !strings.HasPrefix(topic, "response.") {
		return topic
	}
	tokens := strings.Split(topic, ".")
	topic = tokens[0] + "." + strings.Join(tokens[len(tokens)-2:], ".")
	return topic
}

func topic2Exchange(topic string) (string, error) {
	switch {
	case strings.HasPrefix(topic, "data/"):
		return ExchangeNameData, nil
	case strings.HasPrefix(topic, "signal/"):
		return ExchangeNameSignals, nil
	case strings.HasPrefix(topic, "command/"):
		return ExchangeNameRpc, nil
	case strings.HasPrefix(topic, "response/"):
		return ExchangeNameRpc, nil
	case strings.HasPrefix(topic, "log"):
		return ExchangeNameLog, nil
	//case topic == "":
	//	return ExchangeNameTest, nil
	default:
		return "", errors.New("topic cannot be parsed: " + topic)
	}
}

func messageType2Exchange(messageType m2cp.MessageType) (string, error) {
	switch messageType {
	case m2cp.MessageTypeData:
		return ExchangeNameData, nil
	case m2cp.MessageTypeSignal:
		return ExchangeNameSignals, nil
	case m2cp.MessageTypeCommand:
		return ExchangeNameRpc, nil
	case m2cp.MessageTypeResponse:
		return ExchangeNameRpc, nil
	default:
		return "", fmt.Errorf("can't get exchange name for message type: %v", messageType)
	}
}
