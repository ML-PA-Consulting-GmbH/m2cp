package device

import (
	"errors"
	"fmt"
	"m2cp"
	"m2cp/snapd"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type OsVersion struct {
	Apps map[string]uint
}

// This list may be extended, but existing entries never changed
var osVersionDictionary = map[string]string{
	"snapd":             "sd",
	"core":              "16",
	"core18":            "18",
	"core20":            "20",
	"core22":            "22",
	"core24":            "24",
	"m2cp-logstat":      "ls",
	"m2cp-gateway":      "gw",
	"m2cpd":             "md",
	"m2cp-message-hub":  "mh",
	"m2cp-sil0-gadget":  "g0",
	"m2cp-sil0-kernel":  "k0",
	"m2cp-chrony":       "ch",
	"m2cp-coap":         "co",
	"m2cp-lte-coap":     "lc",
	"m2cp-systest":      "st",
	"m2cp-tools":        "tl",
	"m2cp-coap-knorr":   "x0",
	"m2cp-config-knorr": "x1",
}

func osVersionDictionaryLookup(short string) (string, bool) {
	for long, s := range osVersionDictionary {
		if s == short {
			return long, true
		}
	}
	return "", false
}

func decodeOsVersion2(version string) (*OsVersion, error) {
	const errorTemplate = "invalid os-version-v2 string: "

	tokens := strings.Split(version, "-")
	if len(tokens) < 2 {
		return nil, errors.New(errorTemplate + "too few tokens")
	}

	tokens = tokens[1:]
	result := OsVersion{Apps: make(map[string]uint)}

	for i, token := range tokens {
		if len(token) < 3 {
			return nil, fmt.Errorf(errorTemplate+"token %d too short", i)
		}
		snapShort := token[0:2]
		rev, err := strconv.Atoi(token[2:])
		if err != nil {
			return nil, fmt.Errorf(errorTemplate+"can't decode revision of snapShort %s: %s", snapShort, err)
		}

		if snapLong, ok := osVersionDictionaryLookup(snapShort); !ok {
			return nil, fmt.Errorf(errorTemplate+"unkown abbreviation '%s'", snapShort)
		} else {
			result.Apps[snapLong] = uint(rev)
		}
	}

	return &result, nil
}

func encodeOsVersion2(osVersion *OsVersion) string {
	var tokens []string
	for snap, rev := range osVersion.Apps {
		if snapShort, ok := osVersionDictionary[snap]; ok {
			tokens = append(tokens, fmt.Sprintf("%s%d", snapShort, rev))
		}
	}

	sort.Strings(tokens)

	// prefix "2-" signifies version 2 of the encoding
	return "2-" + strings.Join(tokens, "-")
}

func getOsVersion2(ctp m2cp.ContextPlus) (string, error) {
	installed, err := snapd.GetSnapsInstalled(ctp)
	if err != nil {
		return "", fmt.Errorf("could not get installed snaps: %w", err)
	}

	model, err := snapd.GetModelAssertion(ctp)
	if err != nil {
		return "", fmt.Errorf("could not get model assertion: %w", err)
	}

	snapsInstalled := make(map[string]uint)
	for _, snap := range installed {
		revision, err := strconv.Atoi(snap.Revision)
		if err != nil {
			return "", fmt.Errorf("could not parse revision of snap %s: %w", snap.Name, err)
		}
		snapsInstalled[snap.Name] = uint(revision)
	}

	return calculateOsVersion2(ctp, snapsInstalled, *model)

}

func calculateOsVersion2(ctp m2cp.ContextPlus, snaps map[string]uint, model snapd.ModelAssertion) (string, error) {
	coreApps, err := getCoreApps(ctp, model)
	if err != nil {
		return "", fmt.Errorf("could not get core apps: %w", err)
	}

	snapsRelevant := make(map[string]uint)
	for snap, rev := range snaps {
		if slices.Contains(coreApps, snap) {
			snapsRelevant[snap] = rev
		}
	}

	osVersion := OsVersion{Apps: snapsRelevant}
	return encodeOsVersion2(&osVersion), nil

}

func getCoreApps(ctp m2cp.ContextPlus, model snapd.ModelAssertion) ([]string, error) {
	coreApps := []string{
		"snapd",
		"core",
		"core18",
		"core20",
		"core22",
		"core24",
		"core26",
	}

	add := func(s string) {
		if slices.Contains(coreApps, s) {
			return
		}
		coreApps = append(coreApps, s)
	}

	for _, snap := range model.Snaps {
		if _, ok := osVersionDictionary[snap.Name]; ok {
			add(snap.Name)
		} else {
			ctp.LogWarn("unknown core snap %s", snap.Name)
		}
	}

	return coreApps, nil
}
