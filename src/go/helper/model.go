package helper

import (
	"fmt"
	"golang.org/x/net/context"
	gql "m2cpcli/graphql"
	"strings"
)

func GetModelAssertionsByModelName(ctx context.Context, modelName string) ([]*gql.Assertion, error) {
	assertions, err := gql.AssertionsByTypeAndStringFilter(ctx, "model", fmt.Sprintf("model: %s", modelName))
	if err != nil {
		return nil, err
	}

	if len(assertions) == 0 {
		return nil, fmt.Errorf("no model assertions found for model %s", modelName)
	}

	return assertions, nil
}

func GetModelAssertionByModelNameAndRevision(ctx context.Context, modelName string, revision int) (*gql.Assertion, error) {
	assertions, err := GetModelAssertionsByModelName(ctx, modelName)
	if err != nil {
		return nil, err
	}

	for _, assertion := range assertions {
		if assertion.Revision == revision && strings.Contains(assertion.AssertionBody, "model: "+modelName+"\n") {
			return assertion, nil
		}
	}

	return nil, fmt.Errorf("assertion not found for model %s and revision %d", modelName, revision)
}

func GetModelAssertionByModelNameAndLatestRevision(ctx context.Context, modelName string) (*gql.Assertion, error) {
	assertions, err := GetModelAssertionsByModelName(ctx, modelName)
	if err != nil {
		return nil, err
	}

	latestRevision := 0
	var latestAssertion *gql.Assertion
	for _, assertion := range assertions {
		if assertion.Revision > latestRevision {
			latestRevision = assertion.Revision
			latestAssertion = assertion
		}
	}

	if latestAssertion == nil {
		return nil, fmt.Errorf("no model assertion found for model %s", modelName)
	}

	return latestAssertion, nil
}
