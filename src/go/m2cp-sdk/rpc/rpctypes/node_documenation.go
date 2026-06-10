package rpctypes

type NodeDocumentation []RpcDocCommand

type RpcDocCommand struct {
	Name               string                      `json:"name"`
	Type               string                      `json:"type"`
	Description        string                      `json:"description"`
	DescriptionReturns string                      `json:"descriptionReturns"`
	Parameters         map[string]RpcDocParameters `json:"parameters"`
	Returns            map[string]RpcDocReturns    `json:"returns"`
}

type RpcDocParameters struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Optional    bool   `json:"optional"`
	Description string `json:"description"`
}

type RpcDocReturns RpcDocParameters
