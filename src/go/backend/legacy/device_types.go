package legacy

import (
	"strings"
)

// Device type UUIDs. They are equal in all app stores.
const (
	EdgeDeviceTypeId     = "8ca69cde-2b48-4aa0-a0d2-422f63ab3f75"
	RealTimeDeviceTypeId = "9b1c3b9e-e61b-499e-8c95-ff42c7556f29"
)

// Device type name mappings
var deviceTypeAliases = map[string]string{
	// Edge device aliases
	"m2cp":       EdgeDeviceTypeId,
	"edge":       EdgeDeviceTypeId,
	"edgedevice": EdgeDeviceTypeId,

	// Real-time device aliases
	"rtd":            RealTimeDeviceTypeId,
	"realtime":       RealTimeDeviceTypeId,
	"realtimedevice": RealTimeDeviceTypeId,
	"real-time":      RealTimeDeviceTypeId,
}

// GetDeviceTypeIdByName maps a device type name/alias to its corresponding UUID
func GetDeviceTypeIdByName(deviceTypeName string) (string, bool) {
	// Normalize the input (lowercase, remove spaces and dashes for comparison)
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(deviceTypeName, " ", ""), "-", ""))

	if typeId, exists := deviceTypeAliases[normalized]; exists {
		return typeId, true
	}

	return "", false
}

// GetDeviceTypeNameById returns a human-readable name for a device type ID
func GetDeviceTypeNameById(deviceTypeId string) string {
	switch deviceTypeId {
	case EdgeDeviceTypeId:
		return "EdgeDevice"
	case RealTimeDeviceTypeId:
		return "RealTimeDevice"
	default:
		return "Unknown"
	}
}
