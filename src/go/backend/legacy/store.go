package legacy

import (
	"context"
	"errors"
	"fmt"
	"m2cpcli/tools"
)

func GetStoreOwner(ctx context.Context) (tenantName, tenantId string, err error) {
	// TODO: this is just a heuristic due to a missing "get store owner" GQL endpoint.. Replace with a proper function in the future
	arch := Architecture("ARM64")
	app, err := getAppTenantByNameAndArchitecture(ctx, tools.StrPtr("snapd"), &arch)
	if err != nil {
		return "", "", fmt.Errorf("can't fetch store owner details, as snapd(arm64) wasn't found in the App Store: %s", err)
	}
	if app.Apps != nil && app.Apps.Items != nil && len(app.Apps.Items) == 1 {
		tenant := app.Apps.Items[0].Tenant
		return tenant.TenantId, tenant.Alias, nil
	}
	return "", "", fmt.Errorf("invalid gql result when querying for store owner")
}

func GetAssetsOwner(ctx context.Context) (tenantName, tenantId string, err error) {
	// TODO: this is another heuristic to solve the problem: under which tenant should we create assets for Systems
	res, err := BackendInfo(ctx)
	if err != nil {
		return "", "", err
	}
	if res == nil {
		return "", "", errors.New("unexpected empty result")
	}
	if len(res.Tenants) == 1 {
		return res.Tenants[0].TenantAlias, res.Tenants[0].Id, nil
	}
	if len(res.Tenants) > 2 {
		return "", "", errors.New("multi-tenancy is not fully supported in m2cp CLI in this version")
	}
	for _, item := range res.Tenants {
		if item.TenantAlias != "mlpa" {
			return item.TenantAlias, item.Id, nil
		}
	}
	return "", "", errors.New("this shouldn't happen")
}
