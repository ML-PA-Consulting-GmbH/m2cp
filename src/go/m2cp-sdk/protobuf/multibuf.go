package protobuf

import (
	"bytes"
	"fmt"
	"io"
)

type WrappedProtobuf struct {
	Oid      uint
	Protobuf []byte
}

// UnpackWrappedProtobufs parses a CoAP payload into a slice of CoAP data messages.
func UnpackWrappedProtobufs(input []byte) ([]WrappedProtobuf, error) {
	const parserError = "data truncated while reading payload data"

	r := bytes.NewReader(input)
	var messages []WrappedProtobuf

	for r.Len() > 0 {
		oid, _, err := readVarInt(r)
		if err != nil {
			return nil, fmt.Errorf(parserError)
		}

		payloadSize, _, err := readVarInt(r)
		if err != nil {
			// some payload variants send a checksum after the payload - this would be read as "oid" in our loop, followed by an EOF for payloadSize
			// we can therefore ignore the EOF here and finish parsing successfully
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf(parserError)
		}

		if payloadSize > r.Len() {
			return nil, fmt.Errorf("payload size too big: %d", payloadSize)
		}
		if payloadSize < 0 {
			return nil, fmt.Errorf("payload size negative: %d", payloadSize)
		}

		payload := make([]byte, payloadSize)
		if r.Len() < payloadSize {
			return nil, fmt.Errorf(parserError)
		}
		_, err = r.Read(payload)
		if err != nil {
			return nil, fmt.Errorf(parserError)
		}

		messages = append(messages, WrappedProtobuf{
			Oid:      uint(oid),
			Protobuf: payload,
		})
	}

	return messages, nil
}

func PackWrappedProtobufs(messages []WrappedProtobuf) ([]byte, error) {
	var buf bytes.Buffer
	for _, msg := range messages {
		if err := writeVarInt(&buf, int(msg.Oid)); err != nil {
			return nil, err
		}
		if err := writeVarInt(&buf, len(msg.Protobuf)); err != nil {
			return nil, err
		}
		if _, err := buf.Write(msg.Protobuf); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// readVarInt reads a variable length integer from the buffer.
func readVarInt(r *bytes.Reader) (int, int, error) {
	if r == nil {
		return 0, 0, fmt.Errorf("invalid reader")
	}
	var result int32
	var bytesRead int
	buf := make([]byte, 1)

	for {
		if r.Len() == 0 {
			return 0, bytesRead, io.EOF
		}
		if _, err := r.Read(buf); err != nil {
			return 0, bytesRead, err
		}
		bytesRead++
		b := buf[0]
		result |= int32(b&0x7F) << (7 * (bytesRead - 1))
		if b&0x80 == 0 {
			break
		}
	}

	return int(result), bytesRead, nil
}

// writeVarInt writes a variable length integer to the buffer.
func writeVarInt(buf *bytes.Buffer, value int) error {
	for {
		b := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			b |= 0x80
		}
		if err := buf.WriteByte(b); err != nil {
			return err
		}
		if value == 0 {
			break
		}
	}
	return nil
}
