package rpc

import (
	"encoding/json"
	"fmt"
	"m2cp"
	"strings"
	"unicode"
)

type commandDocumentation struct {
	Name               string                       `json:"name"`
	CommandType        string                       `json:"type"`
	Description        string                       `json:"description"`
	DescriptionReturns string                       `json:"descriptionReturns"`
	Parameters         map[string]m2cp.RpcParameter `json:"parameters"`
	Returns            map[string]m2cp.RpcParameter `json:"returns"`
}

func newCommandDocumentation(name, description, descriptionReturns string, c *command) (*commandDocumentation, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil, fmt.Errorf("command '%s' is missing description", name)
	}
	descriptionReturns = strings.TrimSpace(descriptionReturns)
	if descriptionReturns == "" && len(c.returnParams) > 0 {
		return nil, fmt.Errorf("command '%s' is missing description for return parameters", name)
	}

	doc := &commandDocumentation{
		Name:               name,
		Description:        description,
		DescriptionReturns: descriptionReturns,
	}

	for _, param := range c.callParams {
		if !isCamelCase(param.GetName()) {
			return nil, fmt.Errorf("command '%s' call param '%s' is not camel case", name, param.GetName())
		}
		if param.GetDescription() == "" {
			return nil, fmt.Errorf("command '%s' call param '%s' is missing description", name, param.GetName())
		}
		if doc.Parameters == nil {
			doc.Parameters = map[string]m2cp.RpcParameter{}
		}
		doc.Parameters[param.GetName()] = param
	}

	for _, param := range c.returnParams {
		if !isCamelCase(param.GetName()) {
			return nil, fmt.Errorf("command '%s' return param '%s' is not camel case", name, param.GetName())
		}
		if param.GetDescription() == "" {
			return nil, fmt.Errorf("command '%s' return param '%s' is missing description", name, param.GetName())
		}
		if doc.Returns == nil {
			doc.Returns = map[string]m2cp.RpcParameter{}
		}
		doc.Returns[param.GetName()] = param
	}

	return doc, nil
}

func (cd commandDocumentation) GetName() string {
	return cd.Name
}

func (cd commandDocumentation) ToJson() []byte {
	b, _ := json.Marshal(cd)
	return b
}

func isCamelCase(input string) bool {
	return input != "" && !unicode.IsUpper(rune(input[0])) && !strings.Contains(input, "_") && !strings.Contains(input, "-")
}
