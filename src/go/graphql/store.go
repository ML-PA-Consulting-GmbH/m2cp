package graphql

import (
	"context"
)

// StoreAuthorizationEndpoint returns information on login methods available for that store
func StoreAuthorizationEndpoint(ctx context.Context) (string, error) {
	queryString := `query { webLoginOptions {
	  authorizationEndpoint
	}}`

	client, req := PrepareClientAndRequest(ctx, queryString)

	// Open access endpoint - remove authorization header to avoid sending invalidated old token
	req.Header.Del("Authorization")

	var result struct {
		WebLoginOptions struct {
			AuthorizationEndpoint string `json:"authorizationEndpoint"`
		} `json:"webLoginOptions"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return "", err
	}

	return result.WebLoginOptions.AuthorizationEndpoint, nil
}

// CurrentUserToken returns a JWT that contains a "tenant_id" claim. It is only required to fetch this when using
// Browser-Based Authentication.
func CurrentUserToken(ctx context.Context) (string, error) {
	queryString := `query{
  currentUserToken {
    token
  }
}`

	client, req := PrepareClientAndRequest(ctx, queryString)

	var result struct {
		CurrentUserToken struct {
			Token string `json:"token"`
		} `json:"currentUserToken"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return "", err
	}

	return result.CurrentUserToken.Token, nil
}

func VirtualDeviceContainerRegistryCredentials(ctx context.Context) (*VirtualDeviceContainerRegistryCredentialsOutput, error) {
	queryString := `query {
  virtualDeviceContainerRegistryCredentials {
	containerRegistryUri
	username
	password
  }
}`

	client, req := PrepareClientAndRequest(ctx, queryString)

	var result struct {
		VirtualDeviceContainerRegistryCredentials VirtualDeviceContainerRegistryCredentialsOutput `json:"virtualDeviceContainerRegistryCredentials"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return &result.VirtualDeviceContainerRegistryCredentials, nil
}
