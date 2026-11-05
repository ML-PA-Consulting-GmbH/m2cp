package structs

import (
	"fmt"
	"strings"
)

// Note, these AssetModelTypeIds are the same in every store.
const (
	AssetModelTypeIdEd     = "a14e3818-48f4-4474-b48c-6e9a1f0a9d7b"
	AssetModelTypeIdRtd    = "c93bf8b3-ccf2-41c3-8fa6-c6d409f00685"
	AssetModelTypeIdSystem = "bf52781e-ff5c-49a7-ad43-f6e091af3d37"
)

// TODO: these strings may change?
const (
	AssetModelTypeNameEd     = "Edge Device"
	AssetModelTypeNameRtd    = "Real Time Device"
	AssetModelTypeNameSystem = "System"
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
	AssetTypeId            string  `json:"assetType"`
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
	TypeId      string  `json:"typeId"`
	TenantName  string  `json:"tenantName,omitempty"`
	TenantId    string  `json:"tenantId,omitempty"`
}

//func AssetModelTypeIdToName(assetModelTypeId string) string {
//	if assetModelTypeId == AssetModelTypeIdEd {
//		return AssetModelTypeNameEd
//	} else if assetModelTypeId == AssetModelTypeIdRtd {
//		return AssetModelTypeNameRtd
//	} else if assetModelTypeId == AssetModelTypeIdSystem {
//		return AssetModelTypeNameSystem
//	} else {
//		return "Invalid"
//	}
//}

func AssetModelTypeNameToId(assetModelTypeName string) (string, error) {
	if assetModelTypeName == AssetModelTypeNameEd || strings.ToLower(assetModelTypeName) == "ed" {
		return AssetModelTypeIdEd, nil
	} else if assetModelTypeName == AssetModelTypeNameRtd || strings.ToLower(assetModelTypeName) == "rtd" {
		return AssetModelTypeIdRtd, nil
	} else if assetModelTypeName == AssetModelTypeNameSystem || strings.ToLower(assetModelTypeName) == "sys" {
		return AssetModelTypeIdSystem, nil
	} else {
		return "", fmt.Errorf("unknown asset model type '%s'", assetModelTypeName)
	}
}
