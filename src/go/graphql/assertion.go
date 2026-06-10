package graphql

import (
	"context"
	"fmt"
)

func SnapAssertionsBySnapDatabaseIdAndRevision(ctx context.Context, snapDbId UUID, revision int32) ([]*AssertionPlain, error) {
	queryString := `query GetAssertion($id: UUID!, $revision: Int!){
snapAssertions(snap: {snapDeclarationId: $id, revision: $revision}) {
	assertions {
	  assertion
	  revision
	}
  }
}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("id", snapDbId)
	req.Var("revision", revision)

	var result struct {
		SnapAssertions struct {
			Assertions []*AssertionPlain `json:"assertions"`
		} `json:"snapAssertions"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	if len(result.SnapAssertions.Assertions) == 0 {
		return nil, fmt.Errorf("no assertions returned from server")
	}
	return result.SnapAssertions.Assertions, nil
}

// StoreAssertions returns all store related assertions for the users tenant.
func StoreAssertions(ctx context.Context) ([]*AssertionPlain, error) {
	queryString := `query GetStoreAssertions(){
storeAssertionsV2() {
	assertions {
	  assertion
	  revision
	}
  }
}`

	client, req := PrepareClientAndRequest(ctx, queryString)

	var result struct {
		StoreAssertions struct {
			Assertions []*AssertionPlain `json:"assertions"`
		} `json:"storeAssertionsV2"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}
	if len(result.StoreAssertions.Assertions) == 0 {
		return nil, fmt.Errorf("no store assertions returned from server")
	}

	return result.StoreAssertions.Assertions, nil
}

// AssertionsByTypeAndStringFilter returns all assertions that match the given type and an optional string filter.
// Types are: model, snap-declaration, snap-revision, serial, account-key, account.
// Type store is not retrievable via this call.
func AssertionsByTypeAndStringFilter(ctx context.Context, assertionType, filter string) ([]*Assertion, error) {
	queryString := `query Assertions($assertionType: String, $filter: String) {
  assertions(where: {
    assertionType: {
      assertionTypeName: {
        eq: $assertionType
      }
    },
    and: {
      assertionBody: {
        contains: $filter
      }
    }
  }) {
    items {
      id
      assertionBody
      createdAt
      revision
      assertionType {
        id
        assertionTypeName
      }
    }
  }
}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("assertionType", assertionType)
	req.Var("filter", filter)

	var result struct {
		Assertions struct {
			Items []*Assertion `json:"items"`
		} `json:"assertions"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return result.Assertions.Items, nil
}

// AssertionByDbId returns the assertion with the given id
// Todo: Not yet tested!
func AssertionByDbId(ctx context.Context, id UUID) (*Assertion, error) {
	queryString := `query Assertion($id: UUID!) {
  assertion(id: $id) {
	id
	assertionBody
	createdAt
	revision
	assertionType {
	  id
	  assertionTypeName
	}
  }
}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("id", id)

	var result struct {
		Assertion *Assertion `json:"assertion"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return result.Assertion, nil
}

func CreateSystemUserAssertion(ctx context.Context, input SystemUserAssertionCreateInput) (*AssertionPlain, error) {
	queryString := `mutation CreateSystemUserAssertion($input: CreateSystemUserAssertionCommandInput!) {
  createSystemUserAssertionV2(input: $input) {
	assertion
	revision
  }
}`
	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("input", input)

	var result struct {
		CreateSystemUserAssertion *AssertionPlain `json:"createSystemUserAssertionV2"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}
	if result.CreateSystemUserAssertion == nil {
		return nil, fmt.Errorf("no system user assertion returned from server")
	}

	return result.CreateSystemUserAssertion, nil
}
