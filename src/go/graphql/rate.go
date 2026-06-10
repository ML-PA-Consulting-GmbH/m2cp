package graphql

import (
	"fmt"
	"golang.org/x/net/context"
	"m2cpcli/config"
)

var snapsByRatingQueryString = `
query SnapsByRating($statusId: UUID!) {
	snapStatus(id: $statusId) {
		name
		description
		isUnrated
		isUsable
		snapRevisions {
			id
	  		uploadMessage
			revision
			snapDownloadSize
			createdAt
			snapVersion
			snapDeclaration {
				    snapName
					snapDeviceArchitecture
					tenantId
					snapSummary
					snapDescription
					snapBase
					snapId
					createdAt
					assertionId
			}
			description
		}
	}
}
`
var snapRatingsQueryString = `
query SnapRatings($take: Int, $skip: Int){
  snapStatuses(take: $take, skip: $skip)
   {
    items {
	  id
      name
      description
    }
  }
}
`

var snapRatingByNameQueryString = `
query SnapRatingByName($take: Int, $skip: Int, $rating: String){
  snapStatuses(take: $take, skip: $skip, where: {
    name: {eq: $rating}
  }){
    items{
      id,
      isUnrated,
      isUsable,
      description,
      name
    }
  }
}`

// SnapsByRating Returns all snaps with the given rating
func SnapsByRating(ctx context.Context, ratingDatabaseId UUID) (*SnapRating, error) {
	client, request := PrepareClientAndRequest(ctx, snapsByRatingQueryString)
	request.Var("statusId", ratingDatabaseId)

	var result struct {
		RatingType SnapRating `json:"snapStatus"`
	}

	err := client.Run(ctx, request, &result)
	if err != nil {
		return nil, err
	}

	return &result.RatingType, nil
}

// SnapRatings Returns all available snap ratings
func SnapRatings(ctx context.Context) (*[]SnapRating, error) {
	client, request := PrepareClientAndRequest(ctx, snapRatingsQueryString)
	request.Var("take", config.GraphQLMaxPageSize)
	request.Var("skip", 0)

	var result struct {
		SnapStatuses SnapStatusesOutput `json:"snapStatuses"`
	}

	err := client.Run(ctx, request, &result)
	if err != nil {
		return nil, err
	}

	return &result.SnapStatuses.SnapRatings, nil
}

func SnapRatingByName(ctx context.Context, name string) (*SnapRating, error) {
	client, request := PrepareClientAndRequest(ctx, snapRatingByNameQueryString)
	request.Var("take", config.GraphQLMaxPageSize)
	request.Var("skip", 0)
	request.Var("rating", name)

	var result struct {
		SnapStatuses SnapStatusesOutput `json:"snapStatuses"`
	}

	err := client.Run(ctx, request, &result)
	if err != nil {
		return nil, err
	}

	if len(result.SnapStatuses.SnapRatings) > 0 {
		return &result.SnapStatuses.SnapRatings[0], nil
	}

	return nil, fmt.Errorf("no snap rating found with name: %s", name)
}
