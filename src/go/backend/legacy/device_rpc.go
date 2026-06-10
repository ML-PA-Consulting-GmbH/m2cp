package legacy

import (
	"context"
	"fmt"
)

func DeviceRpc(ctx context.Context, node, command string, params map[string]string) (*RpcResponse, error) {
	rpcInput := ExecuteRpcInput{
		Address:    node,
		Command:    command,
		Parameters: []*ExecuteRpcParameterInput{},
	}
	for k, v := range params {
		rpcInput.Parameters = append(rpcInput.Parameters, &ExecuteRpcParameterInput{
			Key:   k,
			Value: v,
		})
	}
	res, err := deviceRpc(ctx, &rpcInput)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Result == nil {
		return nil, fmt.Errorf("failed calling rpc: unexpected empty result")
	}
	if len(res.Result.Responses) == 0 {
		return nil, fmt.Errorf("failed calling rpc: unexpected empty list of responses")
	}
	if len(res.Result.Responses) > 1 {
		return nil, fmt.Errorf("failed calling rpc: unexpected more than one response: %d", len(res.Result.Responses))
	}
	response := RpcResponse{
		Error:   res.Result.Responses[0].Error,
		Message: res.Result.Responses[0].Message,
		Result:  map[string]string{},
	}
	for _, resultValue := range res.Result.Responses[0].Results {
		response.Result[resultValue.Key] = resultValue.Value
	}

	return &response, nil
}

type RpcResponse struct {
	Error   int               `json:"error" yaml:"error"`
	Message string            `json:"message" yaml:"message"`
	Result  map[string]string `json:"result" yaml:"result"`
}
