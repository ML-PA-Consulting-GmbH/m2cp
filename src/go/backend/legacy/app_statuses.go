package legacy

import (
	"strings"
)

// App status UUIDs. They are equal in all app stores.
const (
	BrokenAppStatusId       = "a5f7d36e-8d33-4b87-9d15-65c251a9b530"
	DeniedAppStatusId       = "7b4f9e02-3f87-4e7a-b62c-cf9b691f3a2f"
	DeprecatedAppStatusId   = "c42a8579-5d60-4c86-b6e4-01b3ea4e3839"
	EdgeAppStatusId         = "8f3db93a-4973-4c74-98d5-c8c396764743"
	ExperimentalAppStatusId = "fffce8bd-90c1-4c47-b4b3-722a13f5832d"
	StableAppStatusId       = "2a1ec2e6-5f24-43cb-8561-9c85712cb9b5"
	UnratedAppStatusId      = "d8816d58-52f3-4a45-aef8-9c1c7d6e8b9c"
)

// App status name mappings
var appStatusAliases = map[string]string{
	"broken":       BrokenAppStatusId,
	"denied":       DeniedAppStatusId,
	"deprecated":   DeprecatedAppStatusId,
	"edge":         EdgeAppStatusId,
	"experimental": ExperimentalAppStatusId,
	"stable":       StableAppStatusId,
	"unrated":      UnratedAppStatusId,
}

// GetAppStatusIdByName maps an app status name/alias to its corresponding UUID
func GetAppStatusIdByName(appStatusName string) (string, bool) {
	// Normalize the input (lowercase, remove spaces and dashes for comparison)
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(appStatusName, " ", ""), "-", ""))

	if statusId, exists := appStatusAliases[normalized]; exists {
		return statusId, true
	}

	return "", false
}

// GetAppStatusNameById returns a human-readable name for an app status ID
func GetAppStatusNameById(appStatusId string) string {
	switch appStatusId {
	case BrokenAppStatusId:
		return "Broken"
	case DeniedAppStatusId:
		return "Denied"
	case DeprecatedAppStatusId:
		return "Deprecated"
	case EdgeAppStatusId:
		return "Edge"
	case ExperimentalAppStatusId:
		return "Experimental"
	case StableAppStatusId:
		return "Stable"
	case UnratedAppStatusId:
		return "Unrated"
	default:
		return "Unknown"
	}
}
