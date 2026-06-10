package rpctypes

import (
	"m2cp"
)

func (s *TestSuite) TestCommandResult() {
	result := NewRpcResultSuccess("test")
	s.NotNil(result)
	s.Equal("test", result.GetMessage())
	s.True(result.IsSuccessful())
	s.Equal(m2cp.RpcErrorCodeSuccess, result.GetErrorCode())
}

func (s *TestSuite) TestCommandResultError() {
	result := NewRpcResultError("test", m2cp.RpcErrorCodeInsufficientStorage)
	s.NotNil(result)
	s.Equal("test", result.GetMessage())
	s.False(result.IsSuccessful())
	s.Equal(m2cp.RpcErrorCodeInsufficientStorage, result.GetErrorCode())
	result.SetErrorCode(m2cp.RpcErrorCodeGatewayTimeout)
	s.Equal(m2cp.RpcErrorCodeGatewayTimeout, result.GetErrorCode())
	result.SetSuccess(true)
	s.True(result.IsSuccessful())
	result.SetSuccess(false)
	s.False(result.IsSuccessful())
	result.SetSuccess(true)
	s.True(result.IsSuccessful())
	s.Equal(m2cp.RpcErrorCodeSuccess, result.GetErrorCode())
	result.SetMessage("happy")
	s.Equal("happy", result.GetMessage())
	result.SetInt("x", 7)
	s.Equal("7", result.GetRaw("x", "-1"))
	result.SetLong("x", 14)
	s.Equal("14", result.GetRaw("x", "-1"))
}
