package infrastructure

import (
	"context"
	"fmt"
	gql "m2cpcli/graphql"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var (
	TaskId = ""
)

type authenticatedTransport struct {
	wrapped http.RoundTripper
}

func (transport *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	jwtToken := viper.GetString("jwt")
	if jwtToken != "" {
		req.Header.Set("Authorization", "bearer "+jwtToken)
	}

	if TaskId == "" {
		if gql.TaskId != "" {
			TaskId = gql.TaskId
		} else {
			TaskId = uuid.New().String()
			gql.TaskId = TaskId
		}
	}
	req.Header.Set("X-Task-Id", TaskId)

	return transport.wrapped.RoundTrip(req)
}

func NewGraphqlClient(ctx context.Context) (graphql.Client, error) {
	storeGraphqlEndpoint := viper.GetString("store")

	if storeGraphqlEndpoint == "" {
		return nil, fmt.Errorf("store graphql endpoint is not set")
	}

	httpClient := http.Client{
		Transport: &authenticatedTransport{
			wrapped: http.DefaultTransport,
		},
	}

	graphqlClient := graphql.NewClient(storeGraphqlEndpoint, &httpClient)

	return graphqlClient, nil
}
