package legacy

import (
	"context"
	"fmt"
	"m2cpcli/structs"
	"strconv"
	"strings"
)

func BackendInfo(ctx context.Context) (*structs.Backend, error) {
	res, err := backendInfo(ctx)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("backend info is nil")
	}
	tokens := strings.Split(res.SnapStoreVersion.Version, ".")
	majorVersion, err := strconv.Atoi(tokens[0])
	if err != nil {
		return nil, fmt.Errorf("can't parse major version: %s", err)
	}

	info := structs.Backend{
		Version: structs.BackendVersion{
			Version:      res.SnapStoreVersion.Version,
			MajorVersion: majorVersion,
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
					Id:         module.Id,
					ModuleName: module.ModuleName,
				}
				info.Modules[i].BackendModuleInstances = make([]structs.BackendModuleInstance, len(module.ModuleInstances))
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
