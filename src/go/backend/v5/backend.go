package v5

import (
	"context"
	"fmt"
	"m2cpcli/backend/legacy"
	"m2cpcli/structs"
	"m2cpcli/tools"
	"strconv"
	"strings"
	"time"
)

func BackendInfo(ctx context.Context) (*structs.Backend, error) {
	res, err := backendInfo(ctx)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("backend info is nil")
	}

	tokens := strings.Split(res.StoreInfo.Version, ".")
	majorVersion, err := strconv.Atoi(tokens[0])
	if err != nil {
		return nil, fmt.Errorf("can't parse major version: %s", err)
	}

	info := structs.Backend{
		Version: structs.BackendVersion{
			InstanceId:     &res.StoreInfo.InstanceId,
			RuntimeVersion: &res.StoreInfo.RuntimeVersion,
			StartedAt:      tools.StringToMaybeTime(res.StoreInfo.StartedAt, time.RFC3339),
			Version:        res.SnapStoreVersion.Version,
			MajorVersion:   majorVersion,
		},
		VirtualDeviceContainerRegistryCredentials: &structs.VirtualDeviceContainerRegistryCredentials{
			ContainerRegistryUri: res.VirtualDeviceContainerRegistryCredentials.ContainerRegistryUri,
			Username:             res.VirtualDeviceContainerRegistryCredentials.Username,
			Password:             res.VirtualDeviceContainerRegistryCredentials.Password,
		},
	}

	if res.Modules != nil {
		if len(res.Modules.Items) > 0 {
			info.Modules = make([]structs.BackendModule, len(res.Modules.Items))
			for i, module := range res.Modules.Items {
				info.Modules[i] = structs.BackendModule{
					Id:                     module.Id,
					ModuleName:             module.ModuleName,
					BackendModuleInstances: make([]structs.BackendModuleInstance, len(module.ModuleInstances)),
				}
				for j, instance := range module.ModuleInstances {
					info.Modules[i].BackendModuleInstances[j] = structs.BackendModuleInstance{
						StartedAt:               instance.StartedAt,
						EntryAssemblyModifiedAt: instance.EntryAssemblyModifiedAt,
					}
				}
			}
		}
	}

	if res.StoreSettings != nil && len(res.StoreSettings.Items) > 0 {
		info.Tenants = make([]structs.BackendTenant, len(res.StoreSettings.Items))
		for i, tenant := range res.StoreSettings.Items {
			info.Tenants[i] = structs.BackendTenant{
				Id:          tenant.TenantId,
				TenantName:  tenant.Tenant.TenantName,
				TenantAlias: tenant.Tenant.Alias,
				IsOwner:     tenant.IsOwner,
			}
		}
	}

	return &info, nil
}

// BackendStoreVersion returns the backend's full store version string (e.g.
// "5.2.0"). It errors when the v5 store-info query is unavailable, which is the
// case for legacy (pre-v5) backends - callers should treat that as "older than
// any v5 version".
func BackendStoreVersion(ctx context.Context) (string, error) {
	res, err := backendVersion(ctx)
	if err != nil {
		return "", err
	}
	return res.GetStoreInfo().GetVersion(), nil
}

func BackendMajorVersion(ctx context.Context) (int, error) {
	res, err := backendVersion(ctx)
	if err == nil {
		tokens := strings.Split(res.GetStoreInfo().GetVersion(), ".")
		majorVersion, err := strconv.Atoi(tokens[0])
		if err != nil {
			return 0, fmt.Errorf("can't parse major version: %s", err)
		}
		return majorVersion, nil
	}

	// fallback for legacy
	resLegacy, err := legacy.BackendInfo(ctx)
	if err != nil {
		return 0, err
	}
	return resLegacy.Version.MajorVersion, nil
}
