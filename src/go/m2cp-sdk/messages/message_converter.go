package messages

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/klauspost/compress/zstd"
	"m2cp"
)

var (
	errInsufficientData = errors.New("insufficient data")
	errInvalidFormat    = errors.New("invalid format")
)

func TypeFromJson(messageJson []byte) m2cp.MessageType {
	var m messageEnvelope
	err := json.Unmarshal(messageJson, &m)
	if err != nil {
		return m2cp.MessageTypeInvalid
	}
	return newMessageType(m.Header.Type)
}

func CommandFromJson(messageJson []byte) (m2cp.CommandMessage, error) {
	var m commandMessage
	err := json.Unmarshal(messageJson, &m)
	// m.initAfterParsing (see: C#), used by data-message
	return &m, err
}

func ResponseFromJson(messageJson []byte) (m2cp.ResponseMessage, error) {
	var m responseMessage
	err := json.Unmarshal(messageJson, &m)
	return &m, err
}

func SignalFromJson(messageJson []byte) (m2cp.SignalMessage, error) {
	var m signalMessage
	err := json.Unmarshal(messageJson, &m)
	return &m, err
}

func DataFromJson(messageJson []byte) (m2cp.DataMessage, error) {
	var m dataMessage
	err := json.Unmarshal(messageJson, &m)
	if err != nil {
		return nil, err
	}
	err = m.initAfterParsing()
	return &m, err
}

func MultiMessagePack(individualMessages [][]byte) ([]byte, error) {
	payloadLen := 0
	for i := range individualMessages {
		payloadLen += len(individualMessages[i])
	}
	payloadUncompressed := make([]byte, payloadLen)
	offset := 0
	for i := range individualMessages {
		copy(payloadUncompressed[offset:], individualMessages[i])
		offset += len(individualMessages[i])
	}

	payloadCompressed, err := compressDeflateBytes(payloadUncompressed)
	if err != nil {
		return nil, err
	}

	// allocate a byte array of size payloadLen + 4 * len(individualMessages)
	buffer := make([]byte, 4+1+1+len(payloadCompressed))

	// write the number of payload bytes as the first 4 bytes
	binary.LittleEndian.PutUint32(buffer[:4], uint32(len(payloadCompressed)))

	// Write the format as 5th byte
	buffer[4] = 2

	// Write the compression type as the 6th byte
	buffer[5] = byte(m2cp.MessageSerializerJsonDeflate)

	// Write the individualMessages into the buffer
	copy(buffer[6:], payloadCompressed)

	return buffer, nil
}

func MultiMessageUnpack(packedMessages []byte) (res [][]byte, err error) {
	res, err = multiMessageUnpackFormatPack(packedMessages)
	if err == nil {
		return res, err
	}
	res, err = multiMessageUnpackFormat22(packedMessages)
	if err == nil {
		return res, err
	}
	return multiMessageUnpackFormat1X(packedMessages)
}

func multiMessageUnpackFormat1X(packedMessages []byte) ([][]byte, error) {
	if len(packedMessages) < 5 {
		return nil, errInsufficientData
	}
	switch packedMessages[4] {
	case 1:
		return multiMessageUnpackFormat11(packedMessages)
	case 2:
		return multiMessageUnpackFormat12(packedMessages)
	default:
		return nil, errInvalidFormat
	}
}

func multiMessageUnpackFormat11(packedMessages []byte) ([][]byte, error) {
	const sizeBytes = 4
	const crcBytes = 4
	res := make([][]byte, 0)

	for offset := 0; offset < len(packedMessages); {
		if offset+sizeBytes > len(packedMessages) {
			return nil, errInsufficientData
		}
		payloadLen := int(binary.LittleEndian.Uint32(packedMessages[offset : offset+sizeBytes]))
		if offset+payloadLen+crcBytes > len(packedMessages) {
			return nil, errInsufficientData
		}
		payload := packedMessages[offset : offset+payloadLen+crcBytes]
		res = append(res, payload)
		offset += payloadLen + crcBytes
	}

	return res, nil
}

func multiMessageUnpackFormat12(packedMessages []byte) ([][]byte, error) {
	if len(packedMessages) < 6 {
		return nil, errInsufficientData
	}
	payloadLen := binary.LittleEndian.Uint32(packedMessages[:4])
	payload := packedMessages[6:]

	if packedMessages[5] != byte(m2cp.MessageSerializerJsonDeflate) {
		return nil, errInvalidFormat
	}

	if len(payload) != int(payloadLen) {
		return nil, errInvalidFormat
	}

	payloadUncompressed, err := decompressDeflateBytes(payload)
	if err != nil {
		return nil, err
	}

	offset := 0
	unpackedMessages := make([][]byte, 0)
	for offset < len(payloadUncompressed) {
		if offset+4 > len(payloadUncompressed) {
			return nil, errInsufficientData
		}
		messageLen := int(binary.LittleEndian.Uint32(payloadUncompressed[offset : offset+4]))
		if offset+4+messageLen > len(payloadUncompressed) {
			return nil, errInsufficientData
		}
		unpackedMessages = append(unpackedMessages, payloadUncompressed[offset:offset+4+messageLen])
		offset += 4 + messageLen
	}

	return unpackedMessages, nil
}

// multiMessageUnpackFormatPack unpacks the "pack" format of the m2cp-gateway buffer-service
func multiMessageUnpackFormatPack(packedMessages []byte) ([][]byte, error) {
	if len(packedMessages) < 8 {
		return nil, errInsufficientData
	}
	header := binary.LittleEndian.Uint32(packedMessages[:4])
	if header > 0 {
		return nil, errInvalidFormat
	}
	size := binary.LittleEndian.Uint32(packedMessages[4:8])
	if size == 0 || size > uint32(len(packedMessages)-8) {
		return nil, errInvalidFormat
	}

	return multiMessageUnpackFormat1X(packedMessages[8 : 8+size])
}

func multiMessageUnpackFormat22(packedMessages []byte) ([][]byte, error) {
	if len(packedMessages) < 6 {
		return nil, errInsufficientData
	}

	size := binary.LittleEndian.Uint32(packedMessages[:4])
	format := packedMessages[4]
	compression := packedMessages[5]
	if uint32(len(packedMessages[6:])) != size {
		return nil, errInvalidFormat
	}
	if format != 2 {
		return nil, errInvalidFormat
	}
	if compression != 2 {
		return nil, errInvalidFormat
	}

	decompressed, err := decompressZstd(packedMessages[6:])
	if err != nil {
		return nil, err
	}

	return multiMessageUnpackFormat1X(decompressed)
}

func ToJson(message m2cp.Message) ([]byte, error) {
	return json.Marshal(message)
}

func decompressZstd(data []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	var out []byte
	if out, err = decoder.DecodeAll(data, nil); err != nil && err.Error() != "unexpected EOF" {
		return nil, err
	}
	return out, nil
}
