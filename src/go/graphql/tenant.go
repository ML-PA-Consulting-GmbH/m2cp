package graphql

import (
	"context"
	"fmt"
	"m2cpcli/tools"
)

func TenantIdByAliasOrId(ctx context.Context, tenantAliasOrId string) (UUID, error) {
	var err error
	var tenantId UUID
	if tools.IsValidUuid(tenantAliasOrId) {
		tenantId = UUID(tenantAliasOrId)
	} else {
		alias := tenantAliasOrId
		tenantId, err = TenantIdByAlias(ctx, alias)
		if err != nil {
			return tenantId, err
		}
	}
	return tenantId, nil
}

func TenantIdByAlias(ctx context.Context, tenantAlias string) (UUID, error) {
	queryString := `query SetDefault($tenantAlias: String){
tenants(where: {alias: {eq: $tenantAlias}}, order: {id:ASC}){
	items{
		id
		tenantName
		alias
    }
    totalCount
}}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("tenantAlias", tenantAlias)

	var result struct {
		Tenants TenantCollectionSegment `json:"tenants"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return "", err
	}

	if result.Tenants.TotalCount > 1 {
		return "", fmt.Errorf("tenantAlias \"%s\" is not unique", tenantAlias)
	}
	if result.Tenants.TotalCount == 0 {
		return "", fmt.Errorf("tenantAlias \"%s\" not found", tenantAlias)
	}
	return result.Tenants.Items[0].Id, nil
}

func TenantById(ctx context.Context, tenantId UUID) (*Tenant, error) {
	queryString := `query($id: UUID!){
tenant(id: $id){
	id
	tenantName
	alias
}}`

	client, req := PrepareClientAndRequest(ctx, queryString)
	req.Var("id", tenantId)

	var result struct {
		Item Tenant `json:"tenant"`
	}
	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}

	return &result.Item, nil
}

func Tenants(ctx context.Context) (*[]Tenant, error) {
	queryString := `query Tenants {
  tenants {
    items {
      alias
      id
      tenantName
    }
  }
}`

	client, req := PrepareClientAndRequest(ctx, queryString)

	var result struct {
		TenantList struct {
			Tenants []Tenant `json:"items"`
		} `json:"tenants"`
	}

	err := client.Run(ctx, req, &result)
	if err != nil {
		return nil, err
	}
	return &result.TenantList.Tenants, nil
}
