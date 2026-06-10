package messages

import (
	"fmt"
	"m2cp/rpc/rpctypes"
	"time"
)

func (t *TestSuite) TestResponseMessageToJson() {
	var err error
	from, err := NewAddress(t.ctp, "testnode.appx.testdevice")
	t.NoError(err)
	to, err := NewAddress(t.ctp, "othernode.appy.otherdevice")
	t.NoError(err)

	cmsg, err := NewCommandMessage(from, to, "testCommand")
	t.NoError(err)
	t.NotNil(cmsg)

	cmsg.GetHeader().SetTaskId("a9dc5d85-4d23-4038-91bd-6bf27d01fee6")

	rmsg, err := NewResponseMessage(cmsg)
	t.NoError(err)
	t.NotNil(rmsg)

	rmsg.GetHeader().SetTimestamp(12346.78)
	rmsg.GetHeader().SetId("43")
	rmsg.AddResult(&rpctypes.RpcObject{
		Error:   0,
		Message: "all-good",
		Dict: map[string]string{
			"Foo": "Bar",
		},
	})
	jsonCommand, err := ToJson(rmsg)
	expected := `{"Header":{"Type":"RESPONSE","Topic":"response/testnode.appx.testdevice","Scope":"GLOBAL","Timestamp":"12346.78","Id":"43","TaskId":"a9dc5d85-4d23-4038-91bd-6bf27d01fee6",` +
		`"Origin":"othernode.appy.otherdevice"},"Body":{"CommandId":"` + cmsg.GetHeader().GetId() + `","Responses":[{"Error":0,"Message":"all-good","Results":{"Foo":"Bar"}}]}}`
	t.Equal(expected, string(jsonCommand))
}

func (t *TestSuite) TestCommandResponseTraces() {
	var err error
	from, err := NewAddress(t.ctp, "a.b.x")
	t.NoError(err)
	to, err := NewAddress(t.ctp, "c.d.x")
	t.NoError(err)

	originalCommand, err := NewCommandMessage(from, to, "testCommand")
	t.NoError(err)
	t.NotNil(originalCommand)

	originalCommand.AddTrace("trace1")
	t.ctp.Sleep(100 * time.Millisecond)
	originalCommand.AddTrace("trace2")
	t.ctp.Sleep(100 * time.Millisecond)

	jsonCommand, err := ToJson(originalCommand)
	t.NoError(err)
	t.NotNil(jsonCommand)

	parsedCommand, err := CommandFromJson(jsonCommand)
	t.NoError(err)
	t.NotNil(parsedCommand)

	originalResponse, err := NewResponseMessage(parsedCommand)
	t.NoError(err)
	t.NotNil(originalResponse)

	originalResponse.AddTrace("trace3")
	originalResponse.AddTrace("trace4")

	jsonResponse, err := ToJson(originalResponse)
	t.NoError(err)
	t.NotNil(jsonResponse)

	parsedResponse, err := ResponseFromJson(jsonResponse)
	t.NoError(err)
	t.NotNil(parsedResponse)

	t.Len(parsedResponse.GetTraces(), 4)
	for i, tr := range parsedResponse.GetTraces() {
		t.Equal(fmt.Sprintf("trace%d", i+1), tr.GetRelay())
	}
}
