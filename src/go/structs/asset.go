package structs

import (
	"fmt"
	"strings"
)

const (
	AssetTypeIdEd     = "a14e3818-48f4-4474-b48c-6e9a1f0a9d7b"
	AssetTypeIdRtd    = "c93bf8b3-ccf2-41c3-8fa6-c6d409f00685"
	AssetTypeIdSystem = "bf52781e-ff5c-49a7-ad43-f6e091af3d37"
	AssetTypeEd       = "Edge Device"
	AssetTypeRtd      = "Real Time Device"
	AssetTypeSystem   = "System"
)

type Asset struct {
	Id                     string  `json:"id"`
	Serial                 string  `json:"serial,omitempty"`
	McuId                  *string `json:"mcuId,omitempty"`
	Name                   string  `json:"name,omitempty"`
	Components             []Asset `json:"components"`
	IsSystemOwned          *bool   // this should be false for an actual system, if it's true, then we fetched a child of a system by accident
	SystemAssetId          *string `json:"systemAssetId,omitempty"`
	ParentAssetId          *string `json:"parentAssetId,omitempty"` // systems don't have parent assets, so this should be nil for an actual System ... but its not nil, if we fetched a child of a system by accident
	Description            *string
	AssetType              *string `json:"assetType"`
	AssetModelName         string  `json:"assetModelName,omitempty"`
	AssetModelId           string  `json:"assetModelId,omitempty"`
	AssetModelTypeName     string  `json:"assetModelTypeName,omitempty"`
	AssetModelTypeId       string  `json:"assetModelTypeId,omitempty"`
	AssetModelTypeCategory string  `json:"assetModelTypeCategory,omitempty"`
	TenantName             string  `json:"tenantName,omitempty"`
	TenantId               string  `json:"tenantId,omitempty"`
	DeviceId               *string `json:"deviceId,omitempty"`
	Device                 *Device `json:"device"`
}

type AssetModel struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Revision    *int    `json:"revision"`
	Type        string  `json:"type"`
	TenantName  string  `json:"tenantName,omitempty"`
	TenantId    string  `json:"tenantId,omitempty"`
}

func AssetModelTypeIdToType(assetModelId string) string {
	if assetModelId == AssetTypeIdEd {
		return AssetTypeEd
	} else if assetModelId == AssetTypeIdRtd {
		return AssetTypeRtd
	} else if assetModelId == AssetTypeIdSystem {
		return AssetTypeSystem
	} else {
		return "Invalid"
	}
}

func AssetModelTypeToId(assetModelType string) (string, error) {
	if assetModelType == "Edge Device" || strings.ToLower(assetModelType) == "ed" {
		return AssetTypeIdEd, nil
	} else if assetModelType == "Real Time Device" || strings.ToLower(assetModelType) == "rtd" {
		return AssetTypeIdRtd, nil
	} else if assetModelType == "System" || strings.ToLower(assetModelType) == "sys" {
		return AssetTypeIdSystem, nil
	} else {
		return "", fmt.Errorf("unknown asset model type '%s'", assetModelType)
	}
}
