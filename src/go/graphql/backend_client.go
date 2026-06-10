package graphql

import (
	"context"
	"github.com/ML-PA-Consulting-GmbH/graphql"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

var (
	TaskId = ""
)

func PrepareClientAndRequest(ctx context.Context, queryOrMutation string) (*graphql.Client, *graphql.Request) {
	url := viper.GetString("store")
	jwt := viper.GetString("jwt")

	if TaskId == "" {
		TaskId = uuid.New().String()
	}

	client := graphql.NewClient(url)
	//client.Log = func(s string) {
	//	fmt.Println(s)
	//}
	request := graphql.NewRequest(queryOrMutation)
	request.Header.Set("Authorization", "Bearer "+jwt)
	request.Header.Set("X-Task-Id", TaskId)

	return client, request
}
