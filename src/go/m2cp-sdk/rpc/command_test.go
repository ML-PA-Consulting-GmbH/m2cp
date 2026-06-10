package rpc

import (
	"m2cp"
	"m2cp/rpc/rpctypes"
	"time"
)

func (s *TestSuite) TestCommandMissingFunction() {
	cmd, err := NewRpcCommand("test", "desc", "desc_returns", nil, nil, nil)
	s.Nil(cmd)
	s.Equal("command 'test' is missing a function", err.Error())
}

func (s *TestSuite) TestCommandMissingDescription() {
	cmd, err := NewRpcCommand("test",
		"",
		"desc_returns",
		nil,
		nil,
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			return rpctypes.NewRpcResultSuccess("yeah")
		})
	s.Nil(cmd)
	s.Equal("command 'test' is missing description", err.Error())
}

func (s *TestSuite) TestCommandParamNotCamelCase() {
	cmd, err := NewRpcCommand(
		"test",
		"desc",
		"desc_returns",
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "foo_bar", "bar", false),
		},
		nil,
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			return rpctypes.NewRpcResultSuccess("yeah")
		})
	s.Nil(cmd)
	s.Equal("command 'test' call param 'foo_bar' is not camel case", err.Error())
}

func (s *TestSuite) TestCommandParamWithoutDescription() {
	cmd, err := NewRpcCommand(
		"test",
		"desc",
		"desc_returns",
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "foo", "", false),
		},
		nil,
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			return rpctypes.NewRpcResultSuccess("yeah")
		})
	s.Nil(cmd)
	s.Equal("command 'test' call param 'foo' is missing description", err.Error())
}

func (s *TestSuite) TestCommandMissingDescriptionReturns() {
	// no return params => no description for returns required
	cmd, err := NewRpcCommand(
		"test",
		"desc",
		"",
		nil,
		nil,
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			return rpctypes.NewRpcResultSuccess("yeah")
		})
	s.NotNil(cmd)
	s.Nil(err)
	// return params => description for returns required
	cmd, err = NewRpcCommand(
		"test",
		"desc",
		"",
		nil,
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "foo", "bar", false),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			return rpctypes.NewRpcResultSuccess("yeah")
		})
	s.Nil(cmd)
	s.Equal("command 'test' is missing description for return parameters", err.Error())
}

func (s *TestSuite) TestEmptyFunction() {
	cmd, err := NewRpcCommand(
		"test",
		"desc",
		"desc_returns",
		nil,
		nil, func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) (result m2cp.RpcResult) {
			return rpctypes.NewRpcResultSuccess("yeah")
		})
	s.Equal("test", cmd.GetName())
	s.Nil(err)
	s.NotNil(cmd)
	res := cmd.Execute(s.ctp, nil)
	s.True(res.IsSuccessful())
	s.Equal("yeah", res.GetMessage())
}

func (s *TestSuite) TestCommandDocumentationToJson() {
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
	s.NotNil(cmd)
	json := string(cmd.GetDocumentation().ToJson())
	s.Equal(`{"name":"test","type":"","description":"desc","descriptionReturns":"desc_returns","parameters":{"x":{"name":"x","type":"int","optional":false,"description":"int-param"},"xs":{"name":"xs","type":"[]int","optional":true,"description":"int-array-param"}},"returns":{"d":{"name":"d","type":"double","optional":false,"description":"double-param"},"ds":{"name":"ds","type":"[]double","optional":true,"description":"double-array-param"}}}`, json)

}

func (s *TestSuite) TestCommandExecute() {
	cmd, err := NewRpcCommand(
		"test",
		"desc",
		"desc_returns",
		nil,
		nil,
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			return rpctypes.NewRpcResultSuccess("yeah")
		})
	s.Nil(err)
	s.NotNil(cmd)
	res := cmd.Execute(s.ctp, nil)
	s.True(res.IsSuccessful())
	s.Equal("yeah", res.GetMessage())
}

func (s *TestSuite) TestCommandExecuteClosure() {
	closureValue := "yeah"
	cmd, err := NewRpcCommand(
		"test",
		"desc",
		"desc_returns",
		nil,
		nil,
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			return rpctypes.NewRpcResultSuccess(closureValue)
		})
	s.Nil(err)
	s.NotNil(cmd)
	res := cmd.Execute(s.ctp, nil)
	s.True(res.IsSuccessful())
	s.Equal("yeah", res.GetMessage())
}

func (s *TestSuite) TestCommandExecuteWithParamsAndResults() {
	cmd, err := NewRpcCommand(
		"test",
		"desc",
		"desc_returns",
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "x", "int-param", false),
			NewRpcParameter(m2cp.TypeIntArray, "xs", "int-array-param", false),
			NewRpcParameter(m2cp.TypeDouble, "d", "double-param", false),
			NewRpcParameter(m2cp.TypeDoubleArray, "ds", "double-array-param", false),
			NewRpcParameter(m2cp.TypeBool, "b", "bool-param", false),
			NewRpcParameter(m2cp.TypeDateTime, "dt", "datetime-param", false),
			NewRpcParameter(m2cp.TypeBinary, "bin", "binary-param", false),
			NewRpcParameter(m2cp.TypeJson, "json", "json-param", false),
			NewRpcParameter(m2cp.TypeString, "s", "string-param", true),
		},
		[]m2cp.RpcParameter{
			NewRpcParameter(m2cp.TypeInt, "rx", "int-res", false),
			NewRpcParameter(m2cp.TypeIntArray, "rxs", "int-array-res", false),
			NewRpcParameter(m2cp.TypeDouble, "rd", "double-res", false),
			NewRpcParameter(m2cp.TypeDoubleArray, "rds", "double-array-res", false),
			NewRpcParameter(m2cp.TypeBool, "rb", "bool-res", false),
			NewRpcParameter(m2cp.TypeDateTime, "rdt", "datetime-res", false),
			NewRpcParameter(m2cp.TypeBinary, "rbin", "binary-res", false),
			NewRpcParameter(m2cp.TypeJson, "rjson", "json-res", false),
			NewRpcParameter(m2cp.TypeString, "rs", "string-res", true),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			x := parameters.GetInt("x", 0)
			s.Equal(42, x)
			xs := parameters.GetIntArray("xs")
			s.Equal(2, len(xs))
			d := parameters.GetDouble("d", 0)
			s.Equal(3.25, d)
			ds := parameters.GetDoubleArray("ds")
			s.Equal(2, len(ds))
			s.Equal(1.1, ds[0])
			s.Equal(-2.2, ds[1])
			b := parameters.GetBool("b", false)
			s.True(b)
			dt := parameters.GetDateTime("dt", time.Time{})
			s.Equal(time.Date(2018, 1, 1, 12, 0, 0, 0, time.UTC), dt)
			bin := parameters.GetBinary("bin")
			s.Equal([]byte("hello"), bin)
			type jsonType struct {
				Foo string `json:"foo"`
			}
			var j jsonType
			err := parameters.GetJson("json", &j)
			s.Nil(err)
			s.Equal("bar", j.Foo)

			res := rpctypes.NewRpcResultSuccess("yeah")
			res.SetInt("rx", x+1)
			res.SetIntArray("rxs", []int{3, 4})
			res.SetDouble("rd", d+1)
			res.SetDoubleArray("rds", []float64{2.1, -3.2})
			res.SetBool("rb", !b)
			res.SetDateTime("rdt", dt.Add(time.Hour))
			res.SetBinary("rbin", []byte("world"))
			err = res.SetJson("rjson", jsonType{Foo: "baz"})
			s.NoError(err)
			res.SetString("rs", "bar")
			return res
		})
	s.Nil(err)
	s.NotNil(cmd)

	res := cmd.Execute(s.ctp, map[string]string{
		"x":    "42",
		"xs":   "1,2",
		"d":    "3.25",
		"ds":   "1.1,-2.2",
		"b":    "true",
		"dt":   "2018-01-01T12:00:00Z",
		"bin":  "aGVsbG8=",
		"json": `{"foo":"bar"}`,
		"s":    "bar",
	})
	s.True(res.IsSuccessful())
	s.Equal("yeah", res.GetMessage())
	s.True(res.Has("rs"))

	s.Equal("43", res.GetRaw("rx", ""))
	s.Equal(43, res.GetInt("rx", 0))

	s.True(res.Has("rxs"))
	s.Equal("3,4", res.GetRaw("rxs", ""))
	s.Equal([]int{3, 4}, res.GetIntArray("rxs"))

	s.True(res.Has("rd"))
	s.Equal("4.25", res.GetRaw("rd", ""))
	s.Equal(4.25, res.GetDouble("rd", 0))

	s.True(res.Has("rds"))
	s.Equal("2.1,-3.2", res.GetRaw("rds", ""))
	s.Equal([]float64{2.1, -3.2}, res.GetDoubleArray("rds"))

	s.True(res.Has("rb"))
	s.Equal("false", res.GetRaw("rb", ""))
	s.Equal(false, res.GetBool("rb", true))

	s.True(res.Has("rdt"))
	s.Equal("2018-01-01T13:00:00Z", res.GetRaw("rdt", ""))
	s.Equal(time.Date(2018, 1, 1, 13, 0, 0, 0, time.UTC), res.GetDateTime("rdt", time.Time{}))

	s.True(res.Has("rbin"))
	s.Equal("d29ybGQ=", res.GetRaw("rbin", ""))
	s.Equal([]byte("world"), res.GetBinary("rbin"))

	s.True(res.Has("rjson"))
	s.Equal(`{"foo":"baz"}`, res.GetRaw("rjson", ""))
	var jsonType struct {
		Foo string `json:"foo"`
	}
	s.Nil(res.GetJson("rjson", &jsonType))
	s.Equal("baz", jsonType.Foo)

	s.True(res.Has("rs"))
	s.Equal("bar", res.GetRaw("rs", ""))
	s.Equal("bar", res.GetString("rs", ""))
}
