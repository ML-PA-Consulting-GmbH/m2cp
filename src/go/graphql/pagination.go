package graphql

import (
	"context"
	"github.com/ML-PA-Consulting-GmbH/graphql"
	"m2cpcli/config"
)

type CollectionSegmentInfo struct {
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

type PaginationSegment[T any] struct {
	Items      []T                   `json:"items"`
	TotalCount int                   `json:"totalCount"`
	PageInfo   CollectionSegmentInfo `json:"pageInfo"`
}

func RunPaginated[T any](ctx context.Context, client *graphql.Client, request *graphql.Request) ([]T, error) {
	var queryResult struct {
		PaginationSegment PaginationSegment[T] `json:"result"`
	}

	var accumulatedItems []T
	done := false
	skip := 0
	for !done {
		request.Var("take", config.GraphQLMaxPageSize)
		request.Var("skip", skip)

		err := client.Run(ctx, request, &queryResult)
		if err != nil {
			return nil, err
		}
		accumulatedItems = append(accumulatedItems, queryResult.PaginationSegment.Items...)
		if !queryResult.PaginationSegment.PageInfo.HasNextPage {
			done = true
		} else {
			skip += config.GraphQLMaxPageSize
		}
	}

	// TODO: for some queries this does not hold! But why?
	//if queryResult.PaginationSegment.TotalCount != len(accumulatedItems) {
	//	return nil, fmt.Errorf("could not retrieve all paginated items")
	//}

	return accumulatedItems, nil
}
