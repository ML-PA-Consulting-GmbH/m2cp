package messages

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"m2cp"
	"sync"
)

var crc32Table [16 * 256]uint32
var crc32Initialized = false
var initLock sync.Mutex

func ToBinary(message m2cp.Message, serializer m2cp.MessageSerializer) ([]byte, error) {
	if serializer == m2cp.MessageSerializerProtobuf {
		return ToProtobuf(message)
	}
	jsonMessage, err := ToJson(message)
	if err != nil {
		return nil, err
	}
	return encodeBytes(string(jsonMessage), serializer)
}

func FromBinary(encodedMessage []byte) (m2cp.Message, error) {
	jsonMessage, err := decodeBytes(encodedMessage)
	if err != nil {
		return nil, err
	}
	t := TypeFromJson([]byte(jsonMessage))
	switch t {
	case m2cp.MessageTypeCommand:
		return CommandFromJson([]byte(jsonMessage))
	case m2cp.MessageTypeResponse:
		return ResponseFromJson([]byte(jsonMessage))
	case m2cp.MessageTypeSignal:
		return SignalFromJson([]byte(jsonMessage))
	case m2cp.MessageTypeData:
		return DataFromJson([]byte(jsonMessage))
	default:
		return nil, fmt.Errorf("unknown message type")
	}
}

func CommandFromBinary(encodedMessage []byte) (m2cp.CommandMessage, error) {
	jsonMessage, err := decodeBytes(encodedMessage)
	if err != nil {
		return nil, err
	}
	return CommandFromJson([]byte(jsonMessage))
}

func ResponseFromBinary(encodedMessage []byte) (m2cp.ResponseMessage, error) {
	jsonMessage, err := decodeBytes(encodedMessage)
	if err != nil {
		return nil, err
	}
	return ResponseFromJson([]byte(jsonMessage))
}

func SignalFromBinary(encodedMessage []byte) (m2cp.SignalMessage, error) {
	jsonMessage, err := decodeBytes(encodedMessage)
	if err != nil {
		return nil, err
	}
	return SignalFromJson([]byte(jsonMessage))
}

func DataFromBinary(encodedMessage []byte) (m2cp.DataMessage, error) {
	jsonMessage, err := decodeBytes(encodedMessage)
	if err != nil {
		return nil, err
	}
	return DataFromJson([]byte(jsonMessage))
}

func encodeBytes(json string, serializer m2cp.MessageSerializer) ([]byte, error) {
	compressedJson, err := compress(json, serializer)
	if err != nil {
		return nil, err
	}

	// Required size = 4 bytes size field, 1 byte version, 1 byte compression, n bytes message, 4 bytes CRC32
	requiredSize := 4 + 1 + 1 + len(compressedJson) + 4
	encodedMessage := make([]byte, requiredSize)

	// Write size. Size itself is not included (hence -4)
	binary.LittleEndian.PutUint32(encodedMessage[0:4], uint32(requiredSize-4))

	// Write version
	encodedMessage[4] = 1

	// Write compression type
	encodedMessage[5] = byte(serializer)

	// Write message
	copy(encodedMessage[6:], compressedJson)

	crc := computeCrc32(encodedMessage[4:requiredSize-4], 0, requiredSize-8)

	// Write CRC
	binary.LittleEndian.PutUint32(encodedMessage[requiredSize-4:], crc)

	return encodedMessage, nil
}

func decodeBytes(encodedMessage []byte) (string, error) {
	if len(encodedMessage) < 4 {
		return "", fmt.Errorf("message is too short")
	}

	sizeInMessage := int(binary.LittleEndian.Uint32(encodedMessage[:4]))
	if sizeInMessage != len(encodedMessage)-4 {
		if encodedMessage[0] == '{' && encodedMessage[len(encodedMessage)-1] == '}' {
			return string(encodedMessage), nil
		}
		return "", fmt.Errorf("size field does not match the received length")
	}

	crcInMessage := binary.LittleEndian.Uint32(encodedMessage[len(encodedMessage)-4:])
	crcCalculated := computeCrc32(encodedMessage[4:len(encodedMessage)-4], 0, len(encodedMessage)-8)

	if crcInMessage != crcCalculated {
		return "", fmt.Errorf("CRC verification failed")
	}

	version := encodedMessage[4]
	_ = version
	compression := encodedMessage[5]

	if !isCompressionType(compression) {
		return "", fmt.Errorf("compression type is unknown")
	}

	compressedMessage := encodedMessage[6:sizeInMessage]
	decompressedMessage, err := decompress(compressedMessage, compressionType(compression))
	if err != nil {
		return "", err
	}

	if decompressedMessage == "" {
		return "", fmt.Errorf("decompressed message is empty")
	}

	return decompressedMessage, nil
}

func isCompressionType(compression byte) bool {
	return compression == byte(m2cp.MessageSerializerJson) || compression == byte(m2cp.MessageSerializerJsonDeflate)
}

func compressionType(compression byte) m2cp.MessageSerializer {
	if compression == byte(m2cp.MessageSerializerJson) {
		return m2cp.MessageSerializerJson
	}
	return m2cp.MessageSerializerJsonDeflate
}

func compress(message string, serializer m2cp.MessageSerializer) ([]byte, error) {
	switch serializer {
	case m2cp.MessageSerializerJson:
		return []byte(message), nil
	case m2cp.MessageSerializerJsonDeflate:
		return compressDeflate(message)
	default:
		return nil, fmt.Errorf("compression version %d is unknown", serializer)
	}
}

func compressDeflate(message string) ([]byte, error) {
	var buf bytes.Buffer

	writer, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}

	_, err = io.WriteString(writer, message)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func compressDeflateBytes(message []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	_, err = writer.Write(message)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompress(compressedMessage []byte, serializer m2cp.MessageSerializer) (string, error) {
	switch serializer {
	case m2cp.MessageSerializerJson:
		return string(compressedMessage), nil
	case m2cp.MessageSerializerJsonDeflate:
		return decompressDeflate(compressedMessage)
	default:
		return "", fmt.Errorf("compression type %d is unknown", serializer)
	}
}

func decompressDeflate(compressedMessage []byte) (string, error) {
	messageStream := bytes.NewReader(compressedMessage)
	decompressor := flate.NewReader(messageStream)
	defer decompressor.Close()
	decompressedMessage, err := ioutil.ReadAll(decompressor)
	if err != nil {
		return "", err
	}
	return string(decompressedMessage), nil
}

func decompressDeflateBytes(compressedMessage []byte) ([]byte, error) {
	messageStream := bytes.NewReader(compressedMessage)
	decompressor := flate.NewReader(messageStream)
	defer decompressor.Close()
	decompressedMessage, err := ioutil.ReadAll(decompressor)
	if err != nil {
		return nil, err
	}
	return decompressedMessage, nil
}

func getCrc32(message []byte) []byte {
	return message[len(message)-4:]
}

func computeCrc32(input []byte, offset, length int) uint32 {
	crc := uint32(0)
	initLock.Lock()
	if !crc32Initialized {
		poly := uint32(0xedb88320)
		for i := uint32(0); i < 256; i++ {
			res := i
			for t := 0; t < 16; t++ {
				for k := 0; k < 8; k++ {
					if (res & 1) == 1 {
						res = poly ^ (res >> 1)
					} else {
						res = res >> 1
					}
				}
				crc32Table[(t*256)+int(i)] = res
			}
		}
		crc32Initialized = true
	}
	initLock.Unlock()

	crcLocal := uint32(^crc)
	for length >= 16 {
		a := crc32Table[(3*256)+int(input[offset+12])] ^
			crc32Table[(2*256)+int(input[offset+13])] ^
			crc32Table[(1*256)+int(input[offset+14])] ^
			crc32Table[(0*256)+int(input[offset+15])]

		b := crc32Table[(7*256)+int(input[offset+8])] ^
			crc32Table[(6*256)+int(input[offset+9])] ^
			crc32Table[(5*256)+int(input[offset+10])] ^
			crc32Table[(4*256)+int(input[offset+11])]

		c := crc32Table[(11*256)+int(input[offset+4])] ^
			crc32Table[(10*256)+int(input[offset+5])] ^
			crc32Table[(9*256)+int(input[offset+6])] ^
			crc32Table[(8*256)+int(input[offset+7])]

		d := crc32Table[(15*256)+int(byte(crcLocal)^input[offset])] ^
			crc32Table[(14*256)+int(byte(crcLocal>>8)^input[offset+1])] ^
			crc32Table[(13*256)+int(byte(crcLocal>>16)^input[offset+2])] ^
			crc32Table[(12*256)+int((crcLocal>>24)^uint32(input[offset+3]))]

		crcLocal = d ^ c ^ b ^ a
		offset += 16
		length -= 16
	}

	for length--; length >= 0; length-- {
		crcLocal = crc32Table[byte(crcLocal^uint32(input[offset]))] ^ crcLocal>>8
		offset++
	}

	return crcLocal ^ uint32(0xffffffff)

}
