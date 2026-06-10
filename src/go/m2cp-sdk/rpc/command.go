package rpc

import (
	"fmt"
	"m2cp"
	"m2cp/rpc/rpctypes"
	"unicode"
)

type command struct {
	f             func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult
	callParams    map[string]m2cp.RpcParameter
	returnParams  map[string]m2cp.RpcParameter
	documentation *commandDocumentation
}

func NewRpcCommand(
	name string,
	description string,
	descriptionReturns string,
	callParams, returnParams []m2cp.RpcParameter,
	f func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult,
) (m2cp.RpcCommand, error) {
	var err error
	if f == nil {
		return nil, fmt.Errorf("command '%s' is missing a function", name)
	}

	//if err = validateName(name); err != nil {
	//	return nil, err
	//}

	callParamsByName := make(map[string]m2cp.RpcParameter, len(callParams))
	for _, param := range callParams {
		if _, ok := callParamsByName[param.GetName()]; ok {
			return nil, fmt.Errorf("command '%s' has duplicate call parameter '%s'", name, param.GetName())
		}
		callParamsByName[param.GetName()] = param
	}
	returnParamsByName := make(map[string]m2cp.RpcParameter, len(returnParams))
	for _, param := range returnParams {
		if _, ok := returnParamsByName[param.GetName()]; ok {
			return nil, fmt.Errorf("command '%s' has duplicate return parameter '%s'", name, param.GetName())
		}
		returnParamsByName[param.GetName()] = param
	}

	c := &command{
		callParams:   callParamsByName,
		returnParams: returnParamsByName,
		f:            f,
	}
	c.documentation, err = newCommandDocumentation(name, description, descriptionReturns, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *command) GetName() string {
	return c.documentation.Name
}

func (c *command) GetDocumentation() m2cp.RpcDocumentation {
	return c.documentation
}

func (c *command) Execute(ctp m2cp.ContextPlus, params map[string]string) (res m2cp.RpcResult) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			res = rpctypes.NewRpcResultError(fmt.Sprintf("execution error: %s", r), m2cp.RpcErrorCodeFailed)
		}
	}()
	callParams := &rpctypes.RpcObject{
		Dict:        params,
		Definitions: c.callParams,
	}
	err = callParams.ValidateAll()
	if err != nil {
		return rpctypes.NewRpcResultError(fmt.Sprintf("invalid call parameters: %s", err), m2cp.RpcErrorCodeParametersMismatchDefinition)
	}
	res = c.f(ctp, callParams)
	if res.GetErrorCode() != m2cp.RpcErrorCodeSuccess {
		return res
	}
	if res == nil {
		return rpctypes.NewRpcResultError("execution error: command returned nil", m2cp.RpcErrorCodeFailed)
	}
	res.(*rpctypes.RpcObject).Definitions = c.returnParams
	err = res.ValidateAll()
	if err != nil {
		return rpctypes.NewRpcResultError(fmt.Sprintf("invalid return parameters: %s", err), m2cp.RpcErrorCodeReturnValuesMismatchDefinition)
	}
	return res
}

// validateName checks if the name starts with a capital letter and continues with alphanumerics
func validateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if !unicode.IsUpper(rune(name[0])) {
		return fmt.Errorf("name must start with a capital letter")
	}
	for _, char := range name[1:] {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return fmt.Errorf("name must contain only alphanumeric characters after the first letter")
		}
	}
	return nil
}
