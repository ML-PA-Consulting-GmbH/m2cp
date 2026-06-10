package messages

import (
	"fmt"
	"m2cp"
)

// TestCommandMessage tests the CommandMessage struct and makes sure, that pointers are used correctly.
// Wrongly used pointers cause the "SetX" commands to update copies of the original struct, instead of the original struct.
func (t *TestSuite) TestGetSetHeader() {
	from, err := NewAddress(t.ctp, "testnode.appx.11111111-1111-1111-1111-111111111111")
	t.NoError(err)
	to, err := NewAddress(t.ctp, "othernode.appy.11111111-1111-1111-1111-111111111111")
	t.NoError(err)
	cmsg, err := NewCommandMessage(from, to, "testCommand")
	t.Nil(err)
	t.NotNil(cmsg)
	id := cmsg.GetHeader().GetId()
	h := cmsg.GetHeader()
	h.SetId("42")
	t.Equal("42", cmsg.GetHeader().GetId())
	cmsg.GetHeader().SetId("43")
	t.Equal("43", cmsg.GetHeader().GetId())
	t.NotEqualf(id, cmsg.GetHeader().GetId(), "Id should have changed")
	t.Equal("COMMAND", cmsg.GetHeader().GetTypeString())
	t.Equal(m2cp.MessageTypeCommand, cmsg.GetHeader().GetType())
}

func (t *TestSuite) TestCommandAndResponseMessageToJson() {
	var err error
	from, err := NewAddress(t.ctp, "testnode.appx.00000000-0000-0000-0000-000000000000")
	t.NoError(err)
	to, err := NewAddress(t.ctp, "othernode.appy.11111111-1111-1111-1111-111111111111")
	t.NoError(err)

	parameters := map[string]string{
		"MyInt":    "5",
		"MyFloat":  fmt.Sprintf("%g", 3.141592653589793),
		"MyString": "FooBar",
		"MyEmpty":  "null",
	}
	cmsg, err := NewCommandMessage(from, to, "TestCommand")
	t.Nil(err)
	cmsg.SetParameters([]map[string]string{parameters})
	cmsg.GetHeader().SetTimestamp(12345)
	cmsg.GetHeader().SetId("42")
	cmsg.GetHeader().SetTaskId("a9dc5d85-4d23-4038-91bd-6bf27d01fee6")

	t.Equal(m2cp.MessageScopeGlobal, cmsg.GetHeader().GetScope())

	jsonCommand, err := ToJson(cmsg)
	expected := `{"Header":{"Type":"COMMAND","Topic":"command/othernode.appy.11111111-1111-1111-1111-111111111111","Scope":"GLOBAL","Timestamp":"12345","Id":"42","TaskId":"a9dc5d85-4d23-4038-91bd-6bf27d01fee6","Origin":"testnode.appx.00000000-0000-0000-0000-000000000000"},"Body":{"Command":"TestCommand","Parameters":[{"MyEmpty":"null","MyFloat":"3.141592653589793","MyInt":"5","MyString":"FooBar"}]}}`
	t.Equal(expected, string(jsonCommand))

	parsed, err := CommandFromJson(jsonCommand)
	t.NoError(err)
	t.Equal("TestCommand", parsed.GetCommand())
	t.Equal(m2cp.MessageScopeGlobal, parsed.GetHeader().GetScope())
}

func (t *TestSuite) TestCommandWithRandomParameters() {
	var err error
	from, err := NewAddress(t.ctp, "testnode.appx.00000000-0000-0000-0000-000000000000")
	t.NoError(err)
	to, err := NewAddress(t.ctp, "othernode.appy.11111111-1111-1111-1111-111111111111")
	t.NoError(err)
	cmsg, err := NewCommandMessage(from, to, "TestCommand")
	t.Nil(err)

	psA := t.randomParameters()
	cmsg.SetParameters(psA)

	psB := t.randomParameters()
	for _, p := range psB {
		cmsg.AddParameters(p)
	}

	cmsg.AddParameters(nil)

	psCombined := append(psA, psB...)
	psCombined = append(psCombined, nil)

	cmsgJson, err := ToJson(cmsg)
	t.Nil(err)
	cmsgParsed, err := CommandFromJson(cmsgJson)
	t.Nil(err)
	t.Equal(cmsg.GetHeader().GetId(), cmsgParsed.GetHeader().GetId())
	t.Equal(cmsg.GetHeader().GetTaskId(), cmsgParsed.GetHeader().GetTaskId())
	t.Equal(cmsg.GetHeader().GetTimestamp(), cmsgParsed.GetHeader().GetTimestamp())
	t.Equal(cmsg.GetHeader().GetOrigin(), cmsgParsed.GetHeader().GetOrigin())
	t.Equal(cmsg.GetHeader().GetType(), cmsgParsed.GetHeader().GetType())
	t.Equal(cmsg.GetHeader().GetTopic(), cmsgParsed.GetHeader().GetTopic())
	t.Equal(cmsg.GetCommand(), cmsgParsed.GetCommand())
	t.Equal(cmsg.GetParameters(), cmsgParsed.GetParameters())
	t.Equal(psCombined, cmsgParsed.GetParameters())
}
