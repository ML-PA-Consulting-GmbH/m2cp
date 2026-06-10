package structs

import "time"

type Backend struct {
	Version                                   BackendVersion                             `json:"version"`
	VirtualDeviceContainerRegistryCredentials *VirtualDeviceContainerRegistryCredentials `json:"virtualDeviceContainerRegistryCredentials"`
	Modules                                   []BackendModule                            `json:"modules"`
	Tenants                                   []BackendTenant                            `json:"tenants"`
}

type BackendTenant struct {
	Id          string `json:"id"`
	TenantName  string `json:"name"`
	TenantAlias string `json:"alias"`
	IsOwner     bool   `json:"isOwner"`
}

type BackendModule struct {
	Id                     string                  `json:"id"`
	ModuleName             string                  `json:"name"`
	BackendModuleInstances []BackendModuleInstance `json:"instances"`
}
type BackendModuleInstance struct {
	StartedAt               string `json:"startedAt"`
	EntryAssemblyModifiedAt string `json:"entryAssemblyModifiedAt"`
}

type BackendVersion struct {
	InstanceId     *string    `json:"instanceId"`
	RuntimeVersion *string    `json:"runtimeVersion"`
	StartedAt      *time.Time `json:"startedAt"`
	Version        string     `json:"version"`
	MajorVersion   int        `json:"majorVersion"`
}

type VirtualDeviceContainerRegistryCredentials struct {
	ContainerRegistryUri string `json:"containerRegistryUri"`
	Username             string `json:"username"`
	Password             string `json:"password"`
}

// BackendQueryFilter allows defining filters and sorting for a query
type BackendQueryFilter struct {
	Field    string
	Pattern  *string
	ExactAny []string
	Sort     *BackendQueryFilterSort
	After    *time.Time
	Before   *time.Time
}

type BackendQueryFilterSort string

const (
	BackendQueryFilterArch                    = "arch"
	BackendQueryFilterAssetId                 = "assetId"
	BackendQueryFilterAssetModelTypeId        = "assetModelTypeId"
	BackendQueryFilterAssetName               = "assetName"
	BackendQueryFilterCreatedAt               = "createdAt"
	BackendQueryFilterDeviceLastStoreActivity = "lastAppstoreActivity"
	BackendQueryFilterId                      = "id"
	BackendQueryFilterModifiedAt              = "modifiedAt"
	BackendQueryFilterName                    = "name"
	BackendQueryFilterOSSerial                = "osSerial"
	BackendQueryFilterSortAsc                 = BackendQueryFilterSort("ASC")
	BackendQueryFilterSortDesc                = BackendQueryFilterSort("DESC")
	BackendQueryFilterType                    = "type"
)
