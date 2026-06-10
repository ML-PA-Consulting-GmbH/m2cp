package messages

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"m2cp"
)

func (t *TestSuite) TestSignalMessageToJson() {
	var err error
	senderAddress, err := NewAddressWithSubtopic(t.ctp, "testnode.appx.11111111-1111-1111-1111-111111111111", "testnode/")
	t.NoError(err)
	m, err := NewSignalMessage(senderAddress, "foo", "bar", m2cp.SignalLevelWarning, m2cp.MessageScopeDevice)
	assert.Nil(t.T(), err)
	m.GetHeader().SetTimestamp(12345)
	m.GetHeader().SetId("42")
	m.GetHeader().SetTaskId("a9dc5d85-4d23-4038-91bd-6bf27d01fee6")
	json, err := ToJson(m)
	assert.Nil(t.T(), err)
	expected := `{"Header":{"Type":"SIGNAL","Topic":"signal/testnode/","Scope":"DEVICE","Timestamp":"12345","Id":"42","TaskId":"a9dc5d85-4d23-4038-91bd-6bf27d01fee6","Origin":"testnode.appx.11111111-1111-1111-1111-111111111111"},"Body":{"Type":"WARNING","Name":"foo","Content":"bar"}}`
	assert.Equal(t.T(), expected, string(json))

	parsed, err := SignalFromJson(json)
	t.NoError(err)
	t.Equal(m2cp.SignalLevelWarning, parsed.GetType())
	t.Equal("foo", parsed.GetName())
	t.Equal("bar", parsed.GetContent())
	t.Equal(m2cp.MessageScopeDevice, parsed.GetHeader().GetScope())
}

// TestSignalMessageLegacyScope tests, if legacy messages without scope have the default UNDEFINED scope
func (t *TestSuite) TestSignalMessageLegacyScope() {
	legacyMessage := `{"Header":{"Type":"SIGNAL","Topic":"signal/testnode/","Timestamp":"12345","Id":"42","TaskId":"a9dc5d85-4d23-4038-91bd-6bf27d01fee6","Origin":"testnode.11111111-1111-1111-1111-111111111111"},"Body":{"Type":"WARNING","Name":"foo","Content":"bar"}}`
	parsed, err := SignalFromJson([]byte(legacyMessage))
	t.NoError(err)
	t.Equal(m2cp.MessageScopeUndefined, parsed.GetHeader().GetScope())
}

func (t *TestSuite) TestSignalToBinary() {
	var err error
	senderAddress, err := NewAddressWithSubtopic(t.ctp, "testnode.appx.11111111-1111-1111-1111-111111111111", "testnode/")
	t.NoError(err)
	m, err := NewSignalMessage(senderAddress, "foo", "bar", m2cp.SignalLevelWarning, m2cp.MessageScopeDevice)
	assert.Nil(t.T(), err)
	m.GetHeader().SetTimestamp(12345)
	m.GetHeader().SetId("42")
	m.GetHeader().SetTaskId("a9dc5d85-4d23-4038-91bd-6bf27d01fee6")
	binary, err := ToBinary(m, m2cp.MessageSerializerJsonDeflate)
	// base64 encode
	fmt.Println(base64.StdEncoding.EncodeToString(binary))
}

func (t *TestSuite) TestMultiMessage() {
	amount := 10
	messages := make([]m2cp.Message, amount)
	senderAddress, err := NewAddressWithSubtopic(t.ctp, "testnode.appx.11111111-1111-1111-1111-111111111111", "testnode/")
	t.NoError(err)

	messagesBinary := make([][]byte, amount)

	for i := 0; i < amount; i++ {
		messages[i], err = NewSignalMessage(senderAddress, "foo", fmt.Sprintf("%d", i), m2cp.SignalLevelWarning, m2cp.MessageScopeDevice)
		t.NoError(err)
		var msgBinary []byte
		msgBinary, err = ToBinary(messages[i], m2cp.MessageSerializerJson)
		t.NoError(err)
		messagesBinary[i] = msgBinary
	}

	bytesSingleCompression := 0
	for i := 0; i < amount; i++ {
		comp, err := ToBinary(messages[i], m2cp.MessageSerializerJson)
		t.NoError(err)
		bytesSingleCompression += len(comp)
	}

	packed, err := MultiMessagePack(messagesBinary)
	t.NoError(err)
	t.NotNil(packed)

	fmt.Printf("Single compression: %d bytes\n", bytesSingleCompression)
	fmt.Printf("Multi compression: %d bytes\n", len(packed))

	// print packed message in base64
	fmt.Println(base64.StdEncoding.EncodeToString(packed))

	unpacked, err := MultiMessageUnpack(packed)
	t.NoError(err)
	t.NotNil(unpacked)
	t.Len(unpacked, amount)

	for i := 0; i < amount; i++ {
		assert.Equal(t.T(), messagesBinary[i], unpacked[i])
	}
}
