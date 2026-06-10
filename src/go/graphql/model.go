package graphql

import (
	"context"
	"fmt"
)

type EdgeDeviceModelQueryOptions struct {
	WithAssertion bool
}

func Models(ctx context.Context) ([]Model, error) {
	queryString := `query{
result: models
  {
	items{
      architecture
      name
      revision
    }  
  }
}`
	client, req := PrepareClientAndRequest(ctx, queryString)

	var res struct {
		Result struct {
			Items []Model `json:"items"`
		} `json:"result"`
	}
	err := client.Run(ctx, req, &res)
	if err != nil {
		return nil, err
	}

	return res.Result.Items, nil
}

func EdgeDeviceModelById(ctx context.Context, modelId UUID, options EdgeDeviceModelQueryOptions) (*EdgeDeviceModel, error) {
	queryTemplate := `query EdgeDeviceModelById($modelId: UUID!) {
result: edgeDeviceModels(where: { id: {eq: $modelId}}) {
    items {
      id      
      modelType
      modelName
      modelRevision
      modelArchitecture
	  isTpmRequired
      uploadMessage
      {{ if .WithAssertion }} 
	  modelAssertion {
        assertionBody
      }
	  {{ end }}
    }    
  }
}`
	queryString, err := RenderQueryOrMutation(queryTemplate, options)
	if err != nil {
		return nil, err
	}
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("modelId", modelId)

	var res struct {
		Result struct {
			Items []EdgeDeviceModel `json:"items"`
		} `json:"result"`
	}
	err = client.Run(ctx, req, &res)
	if err != nil {
		return nil, err
	}

	if len(res.Result.Items) == 0 {
		return nil, fmt.Errorf("could not find a matching model")
	}
	if len(res.Result.Items) > 1 {
		return nil, fmt.Errorf("model ID is not unique")
	}
	return &res.Result.Items[0], nil
}

func EdgeDeviceModelIdByTypeNameRevision(ctx context.Context, modelType string, modelName string, modelRevision int) (UUID, error) {
	queryString := `query EdgeDeviceModelIdByTypeNameRevision($modelType: String!, $modelName: String!, $modelRevision: Int!) {          
  result: edgeDeviceModels(
    where: {
      and: [
        {modelType: {eq: $modelType}},
        {modelName: {eq: $modelName}},
        {modelRevision: {eq: $modelRevision}}
      ]
    }
  ) {
    items {
      id
    }    
  }
}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("modelType", modelType)
	req.Var("modelName", modelName)
	req.Var("modelRevision", modelRevision)

	var res struct {
		Result struct {
			Items []EdgeDeviceModel `json:"items"`
		} `json:"result"`
	}
	err := client.Run(ctx, req, &res)
	if err != nil {
		return "", err
	}

	if len(res.Result.Items) == 0 {
		return "", fmt.Errorf("could not find a matching model")
	}
	if len(res.Result.Items) > 1 {
		return "", fmt.Errorf("model ID is not unique")
	}
	return res.Result.Items[0].Id, nil
}

func GetEdgeDeviceModelBridgeSnapDeclaration(ctx context.Context, modelId UUID) ([]EdgeDeviceModelBridgeSnapDeclaration, error) {
	queryString := `query GetEdgeDeviceModelBridgeSnapDeclarations($modelId: UUID!) {
result: edgeDeviceModelBridgeSnapDeclarations(where: { edgeDeviceModelId: {eq: $modelId}}) {
  items {
    id    
    snapDeclarationId
    snapDeclaration {
      id
      snapId
      snapBase
      snapDeviceArchitecture
      snapName
      snapDescription
      snapSummary
      tenantId
      createdAt
      assertionId      
    }
  }
}}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("modelId", modelId)

	var res struct {
		Result struct {
			Items []EdgeDeviceModelBridgeSnapDeclaration `json:"items"`
		} `json:"result"`
	}
	err := client.Run(ctx, req, &res)
	if err != nil {
		return nil, err
	}

	return res.Result.Items, nil
}
