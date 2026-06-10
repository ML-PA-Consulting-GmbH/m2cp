package rpc

import (
	"m2cp"
)

func (s *TestSuite) TestParam() {
	param := NewRpcParameter(m2cp.TypeString, "testParam", "testParamDesc", false)
	s.NotNil(param)
}

func (s *TestSuite) TestNewCommandParameter() {
	dataType := m2cp.TypeString
	name := "testName"
	description := "testDescription"
	const optional = true

	cp := NewRpcParameter(dataType, name, description, optional)
	s.Equal(name, cp.GetName())
	s.Equal(dataType, cp.GetType())
	s.Equal(description, cp.GetDescription())
	s.Equal(optional, cp.IsOptional())
}
