package messages

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
	"m2cp"
	"testing"
)

/**
 * An identical test is implemented in the C# code base to assure compatibility.
 */
func TestComputeCrc32(t *testing.T) {
	const crc32Reference uint32 = 0xD4A1185
	const testString string = "hello world"
	testBytes := []byte(testString)
	crc32 := computeCrc32(testBytes, 0, len(testBytes))
	assert.Equal(t, crc32Reference, crc32)
}

func TestEncodeBinaryUncompressed(t *testing.T) {
	encodeBytesGenericTest(t, m2cp.MessageSerializerJson)
}

func TestEncodeBinaryCompressedDeflate(t *testing.T) {
	encodeBytesGenericTest(t, m2cp.MessageSerializerJsonDeflate)
}

func TestDecodeBytesUncompressed(t *testing.T) {
	decodeBytesGenericTest(t, m2cp.MessageSerializerJson)
}

func TestDecodeBytesCompressedDeflate(t *testing.T) {
	decodeBytesGenericTest(t, m2cp.MessageSerializerJsonDeflate)
}

func TestCrossSDKCompatibilityDecoding(t *testing.T) {
	const encodedCommandMessageCsharp = `sQAAAAEBNY7BDoIwEET/Zc+ggAWEm2KMHhAP3IyHLV1MY0pJISaE9N9tTbzNvMzOzgoXQkEGyhXaZSQooWrq+nA7QQCtHmXnSK/1lqPxRCqaZlSjo3GyY2nm4FU4x3gfJ7SPQyyiPGRFlIZFlPQhFxnPe9xjwRJfgNP7l3e6MfIlB6cHLci1bQR9ZEe+1AZw1GLxqyqtFA7+pHWv/y6AOxpUNJOZoHyscNbabwL7tPYLpR8XdQ==`
	const encodedCommandMessageGolang = `uQAAAAEBNI5PS8QwFMS/y5xTbbPdP81NV0QP63roTTy8bF4kSPpKUoRS8t0lgscZZn78NrwwOU4wG8Z1Zhicr5fLw9sTFEaZww0GXuTeUqpNiJwXijMMOr3r9wcovDoY9NZ3mk9dQ0N7bPqh3TdDq31j3cEePZ1o6HUFUP7+20PhmsJXmGAwieNO7+4c/4QbV2hReBS3VquzxEhTvYycl/+k8E6JIi+cMszHhmeR6oTyWcpvAAAA//87/DSy`
	msgs := []string{
		encodedCommandMessageCsharp,
		encodedCommandMessageGolang,
	}
	for _, msg := range msgs {
		bs, err := base64.StdEncoding.DecodeString(msg)
		assert.NoError(t, err, "Error decoding base64 string")
		m, err := CommandFromBinary(bs)
		assert.NoError(t, err, "Error decoding message")
		assert.Equal(t, "TestCommand", m.GetCommand())
	}
}

func TestFeedDecoderWithJson(t *testing.T) {
	const encodedCommandMessagePlain = `{"Header":{"Type":"COMMAND","Topic":"foo/bar","Timestamp":"123456","Id":"4bf12e81-a907-4905-902f-bd6b7fa8a942","TaskId":"","Origin":"node123.device456"},"Body":{"Command":"TestCommand","Parameters":[{"Foo":"1"}]}}`
	m, err := CommandFromBinary([]byte(encodedCommandMessagePlain))
	assert.NoError(t, err, "Error decoding message")
	assert.Equal(t, "TestCommand", m.GetCommand())
}

func encodeBytesGenericTest(t *testing.T, serializer m2cp.MessageSerializer) {
	json := makeJson("COMMAND", `
	            "Command": "TestCommand",
	            "Parameters": [{
	  	            "Foo":"1"
                }]
            `)
	m, err := CommandFromJson([]byte(json))
	assert.NoError(t, err, "Error parsing json")

	encodedMessage, err := ToBinary(m, serializer)
	assert.NoError(t, err, "Error encoding message")

	fmt.Printf("Encoded message: %v\n", toBase64(encodedMessage))
	fmt.Printf("CRC: %v\n", toBase64(encodedMessage[len(encodedMessage)-4:]))

	// Assert version
	if encodedMessage[4] != 1 {
		t.Errorf("Expected version to be 1, but got %v", encodedMessage[4])
	}
	// Assert compression to be deflate
	if encodedMessage[5] != byte(serializer) {
		t.Errorf("Expected compression to be %v, but got %v", serializer, encodedMessage[5])
	}

	sizeInMessage := binary.LittleEndian.Uint32(encodedMessage[0:4])
	if sizeInMessage != uint32(len(encodedMessage)-4) {
		t.Errorf("Expected sizeInMessage to be %v, but got %v", len(encodedMessage)-4, sizeInMessage)
	}

	crcInMessage := binary.LittleEndian.Uint32(encodedMessage[len(encodedMessage)-4:])
	crcCalculated := computeCrc32(encodedMessage[4:len(encodedMessage)-4], 0, len(encodedMessage)-8)
	if crcCalculated != crcInMessage {
		t.Errorf("Expected crc to be %v, but got %v", crcInMessage, crcCalculated)
	}
}

func toBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func decodeBytesGenericTest(t *testing.T, serializer m2cp.MessageSerializer) {
	json := makeJson("COMMAND", `
	            "Command": "TestCommand",
	            "Parameters": [{
	  	            "Foo":"1"
                }]
            `)
	m, err := CommandFromJson([]byte(json))
	assert.NoError(t, err, "Error parsing json")

	encodedMessage, err := ToBinary(m, serializer)
	assert.NoError(t, err, "Error encoding message")
	decodedMessage, err := CommandFromBinary(encodedMessage)

	if decodedMessage.GetCommand() != "TestCommand" {
		t.Errorf("Expected command to be TestCommand, but got %v", decodedMessage.GetCommand())
	}

	if decodedMessage.GetHeader().GetId() != m.GetHeader().GetId() {
		t.Errorf("Expected id to be %v, but got %v", m.GetHeader().GetId(), decodedMessage.GetHeader().GetId())
	}

	if decodedMessage.GetHeader().GetTopic() != m.GetHeader().GetTopic() {
		t.Errorf("Expected topic to be %v, but got %v", m.GetHeader().GetTopic(), decodedMessage.GetHeader().GetTopic())
	}

	if decodedMessage.GetHeader().GetTimestamp() != m.GetHeader().GetTimestamp() {
		t.Errorf("Expected timestamp to be %v, but got %v", m.GetHeader().GetTimestamp(), decodedMessage.GetHeader().GetTimestamp())
	}
}

func makeJson(messageType string, body string) string {
	return fmt.Sprintf(`
	{
	  "Header": {
		"Type": "%s",
		"Topic": "foo/bar",
		"Timestamp": "123456",
		"Id": "4bf12e81-a907-4905-902f-bd6b7fa8a942",
		"Origin": "node123.device456"
	  },
	  "Body": {%s}}`, messageType, body)
}
