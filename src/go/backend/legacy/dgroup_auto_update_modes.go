package legacy

import (
	"strings"
)

// Auto update mode UUIDs. They are equal in all app stores.
const (
	EdgeAutoUpdateModeId   = "9e794835-3dd8-4982-9241-cdb9a02703eb"
	OffAutoUpdateModeId    = "120f86a6-1204-4e7b-b3bf-11d8a669f37e"
	StableAutoUpdateModeId = "dfb20d44-96b0-41b3-bf87-98dfa3621c60"
)

// Auto update mode name mappings
var autoUpdateModeAliases = map[string]string{
	// Edge mode aliases
	"edge": EdgeAutoUpdateModeId,

	// Off mode aliases
	"off": OffAutoUpdateModeId,

	// Stable mode aliases
	"stable": StableAutoUpdateModeId,
}

// GetAutoUpdateModeIdByName maps an auto update mode name/alias to its corresponding UUID
func GetAutoUpdateModeIdByName(autoUpdateModeName string) (string, bool) {
	// Normalize the input (lowercase, remove spaces and dashes for comparison)
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(autoUpdateModeName, " ", ""), "-", ""))

	if modeId, exists := autoUpdateModeAliases[normalized]; exists {
		return modeId, true
	}

	return "", false
}

// GetAutoUpdateModeNameById returns a human-readable name for an auto update mode ID
func GetAutoUpdateModeNameById(autoUpdateModeId string) string {
	switch autoUpdateModeId {
	case EdgeAutoUpdateModeId:
		return "Edge"
	case OffAutoUpdateModeId:
		return "Off"
	case StableAutoUpdateModeId:
		return "Stable"
	default:
		return "Unknown"
	}
}
