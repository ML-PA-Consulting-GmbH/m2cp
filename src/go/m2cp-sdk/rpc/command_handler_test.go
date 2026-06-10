package rpc

import (
	"fmt"
	"m2cp"
	"m2cp/messages"
	"m2cp/rpc/rpctypes"
)

func (s *TestSuite) TestEmptyCommandHandler() {
	handler, err := NewRpcHandler(nil)
	s.NotNil(handler)
	s.Nil(err)
	docs := handler.GetDocumentation()
	s.NotNil(docs)
	s.Equal(3, len(docs))
}

func (s *TestSuite) TestCommandHandlerDocumentation() {
	cmd, err := NewRpcCommand(
		"test",
		"desc",
		"desc_returns",
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "x", "int-param", false),
			NewRpcParameter(m2cp.TypeIntArray, "xs", "int-array-param", true),
		},
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeDouble, "d", "double-param", false),
			NewRpcParameter(m2cp.TypeDoubleArray, "ds", "double-array-param", true),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			return rpctypes.NewRpcResultSuccess("yeah")
		})
	s.Nil(err)
	handler, err := NewRpcHandler(cmd)
	s.NotNil(handler)
	s.Nil(err)
	docs := handler.GetDocumentation()
	s.NotNil(docs)
	s.Equal(4, len(docs))
	// verify documentation
	for _, doc := range docs {
		if doc.GetName() == "test" {
			s.Equal(`{"name":"test","type":"","description":"desc","descriptionReturns":"desc_returns","parameters":{"x":{"name":"x","type":"int","optional":false,"description":"int-param"},"xs":{"name":"xs","type":"[]int","optional":true,"description":"int-array-param"}},"returns":{"d":{"name":"d","type":"double","optional":false,"description":"double-param"},"ds":{"name":"ds","type":"[]double","optional":true,"description":"double-array-param"}}}`, string(doc.ToJson()))
		}
	}
}

func (s *TestSuite) TestCommandHandlerExecute() {
	var err error
	from, err := messages.NewAddress(s.ctp, "testnode.appx.testdevice")
	s.NoError(err)
	to, err := messages.NewAddress(s.ctp, "othernode.appy.otherdevice")
	s.NoError(err)

	parameters := map[string]string{
		"MyInt":    "5",
		"MyFloat":  fmt.Sprintf("%g", 3.141592653589793),
		"MyString": "FooBar",
		"MyEmpty":  "null",
	}
	cmsg, err := messages.NewCommandMessage(from, to, "testCommand")
	s.Nil(err)
	cmsg.SetParameters([]map[string]string{parameters})
	cmsg.GetHeader().SetTimestamp(12345)
	cmsg.GetHeader().SetId("42")
	cmsg.GetHeader().SetTaskId("a9dc5d85-4d23-4038-91bd-6bf27d01fee6")

	handler, err := NewRpcHandler(exampleCommand())
	s.NotNil(handler)
	s.Nil(err)
	rmsg, err := handler.Process(s.ctp, cmsg)
	s.NoError(err)
	s.NotNil(rmsg)
}

func exampleCommand() m2cp.RpcCommand {
	cmd, _ := NewRpcCommand(
		"test",
		"desc",
		"desc_returns",
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "x", "int-param", false),
			NewRpcParameter(m2cp.TypeIntArray, "xs", "int-array-param", true),
		},
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeDouble, "d", "double-res", false),
			NewRpcParameter(m2cp.TypeInt, "y", "int-res", true),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			res := rpctypes.NewRpcResultSuccess("yeah")
			res.SetDouble("d", 3.141592653589793)
			res.SetInt("y", 42)
			return res
		})
	return cmd
}
