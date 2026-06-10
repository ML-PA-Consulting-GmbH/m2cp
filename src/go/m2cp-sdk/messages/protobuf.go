package messages

import (
	"google.golang.org/protobuf/proto"
	"m2cp"
	"m2cp/messages/pb"
)

func ToProtobuf(message m2cp.Message) ([]byte, error) {

	header := message.GetHeader()
	msg := &pb.Message{
		Header: &pb.Header{
			Id:        header.GetId(),
			Type:      pb.MessageType(header.GetType()),
			Topic:     header.GetTopic(),
			Scope:     pb.MessageScope(header.GetScope()),
			TimeStamp: header.GetTimestamp().Unix(),
			TaskId:    header.GetTaskId(),
			Origin:    header.GetOrigin(),
		},
	}

	// Serialize the Body

	switch t := message.(type) {
	//case:
	case m2cp.SignalMessage:
		_ = t
	// msg is a SignalMessage
	//case *m2cp.DataMessage:
	// msg is a DataMessage
	default:
		// msg is some other type
	}
	//switch message.(type) {
	//case signalMessage:
	//	msg.Body = &pb.Message_CommandBody{CommandBody: convertCommandBody(message.CommandBody)}
	//case m2cp.MessageTypeResponse:
	//	msg.Body = &pb.Message_ResponseBody{ResponseBody: convertResponseBody(message.ResponseBody)}
	//case m2cp.MessageTypeSignal:
	//	msg.Body = &pb.Message_SignalBody{SignalBody: convertSignalBody(message.SignalBody)}
	//case m2cp.MessageTypeData:
	//	msg.Body = &pb.Message_DataBody{DataBody: convertDataBody(message.DataBody)}
	//default:
	//	return nil, errors.New("unsupported message type")
	//}

	binData, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	//return compressData(binData)

	return binData, nil
}
