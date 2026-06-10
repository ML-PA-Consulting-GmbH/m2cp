package rpc

import (
	"m2cp"
)

type rpcParameter struct {
	Name        string    `json:"name"`
	Type        m2cp.Type `json:"-"`
	TypeString  string    `json:"type"`
	Optional    bool      `json:"optional"`
	Description string    `json:"description"`
	Value       string    `json:"-"`
}

func NewRpcParameter(dataType m2cp.Type, name, description string, optional bool) m2cp.RpcParameter {
	return rpcParameter{
		Name:        name,
		Type:        dataType,
		TypeString:  dataType.String(),
		Optional:    optional,
		Description: description,
	}
}

func (cp rpcParameter) GetName() string {
	return cp.Name
}

func (cp rpcParameter) GetType() m2cp.Type {
	return cp.Type
}

func (cp rpcParameter) IsOptional() bool {
	return cp.Optional
}

func (cp rpcParameter) GetDescription() string {
	return cp.Description
}
