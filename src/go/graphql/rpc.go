package graphql

import "context"

func ExecuteRpc(ctx context.Context, rpcInput *ExecuteRpcInput) (*ExecuteRpcOutput, error) {
	queryString := `mutation RunRpc($rpcInput: ExecuteRpcInput!){
  result: executeRPC(
    input: $rpcInput
  ){
    commandId
    responses{
      error
      message
      results{
        key
        value        
      }
    }
  }
}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("rpcInput", rpcInput)

	var result struct {
		Item ExecuteRpcOutput `json:"result"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return &result.Item, nil
}
