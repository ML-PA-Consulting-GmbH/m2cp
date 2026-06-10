package device

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

type osVersionV1 struct {
	Version      int    `yaml:"version" json:"version"`
	SnapstoreUrl string `yaml:"snapstore-url" json:"snapstore-url"`
	Board        string `yaml:"board" json:"board"`
	TenantAlias  string `yaml:"tenant-alias" json:"tenant-alias"`
	Kernel       struct {
		Name     string `yaml:"name" json:"name"`
		Revision int    `yaml:"revision" json:"revision"`
	} `yaml:"kernel" json:"kernel"`
	Gadget struct {
		Name     string `yaml:"name" json:"name"`
		Revision int    `yaml:"revision" json:"revision"`
	} `yaml:"gadget" json:"gadget"`
	SnapdRevision          int `yaml:"snapd-revision" json:"snapd-revision"`
	CoreRevision           int `yaml:"core-revision" json:"core-revision"`
	Core18Revision         int `yaml:"core18-revision" json:"core-18-revision"`
	Core20Revision         int `yaml:"core20-revision" json:"core-20-revision"`
	Core22Revision         int `yaml:"core22-revision" json:"core-22-revision"`
	Core24Revision         int `yaml:"core24-revision" json:"core-24-revision"`
	M2cpGatewayRevision    int `yaml:"m2cp-gateway-revision" json:"m2cp-gateway-revision"`
	M2cpLogStatRevision    int `yaml:"m2cp-logstat-revision" json:"m2cp-logstat-revision"`
	M2cpMessageHubRevision int `yaml:"m2cp-message-hub-revision" json:"m2cp-message-hub-revision"`
}

func getKeyForValue(dictionary map[string]string, value string) (string, error) {
	for key, val := range dictionary {
		if val == value {
			return key, nil
		}
	}
	return "", fmt.Errorf("could not find key for value \"%s\"", value)
}

func getValueForKey(dictionary map[string]string, key string) (string, error) {
	if value, found := dictionary[key]; found {
		return value, nil
	} else {
		return "", fmt.Errorf("could not find value for key \"%s\"", key)
	}
}

// We use a SHA1 hash of the Snapstore URL for privacy reasons. Clients must not know other clients by analysing
// the compiled binary. Calculate the hashes like this and add the plain URL as a comment:
//
//	$ echo -n "https://app-snapstore-api-st02-dev.azurewebsites.net" | sha1sum --text | cut -d " " -f1
var storeCodesV1 = map[string]string{
	"2ddd754c377fc0089ed0e26027f8638af91b621a": "2d", // https://app-snapstore-api-st02-dev.azurewebsites.net
	"a7573aa61dc58d4de4f095f1377454cb7158541b": "2d", // https://app-snapstore-api-st02-test.azurewebsites.net
	"82ced72fdac81fa46d99191d3bf8f282fd2ae75a": "3d", // https://app-snapstore3-gateway-dev.azurewebsites.net/graphql
}
var boardCodesV1 = map[string]string{
	"phyboard-pollux": "pp0",
	"m2cp-sil0":       "s0",
}

// We use a SHA1 hash of the Tenant's Alias for privacy reasons. Clients must not know other clients by analysing
// the compiled binary. Calculate the hashes like this and add the plain URL as a comment:
//
//	$ echo -n "mlpa" | sha1sum --text | cut -d " " -f1
var tenantCodesV1 = map[string]string{
	"ab3178d1351758363af471e892d64b25a9c5abaa": "m",  // mlpa, ML!PA Consulting GmbH
	"c0b138728809433bd8592760d284a46363161a8c": "kb", // knorr, Knorr-Bremse Services GmbH
	"46a808cfd5beafa5e60aefee867bf92025dc2849": "g",  // generic, Generic Tenant
	"a94a8fe5ccb19ba61c4c0873d391e987982fbbd3": "t",  // test, Test Tenant
}

func parseVersion1(versionTokens []string) (*osVersionV1, error) {
	result := osVersionV1{}

	var err error
	result.Version, err = strconv.Atoi(versionTokens[0])
	if err != nil {
		return nil, fmt.Errorf("invalid version: %s", versionTokens[0])
	}
	result.SnapstoreUrl, err = getKeyForValue(storeCodesV1, versionTokens[1])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.Board, err = getKeyForValue(boardCodesV1, versionTokens[2])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.TenantAlias, err = getKeyForValue(tenantCodesV1, versionTokens[3])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}

	result.Kernel.Name = result.Board + "-kernel"
	result.Kernel.Revision, err = strconv.Atoi(versionTokens[4])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.Gadget.Name = result.Board + "-gadget"
	result.Gadget.Revision, err = strconv.Atoi(versionTokens[5])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.SnapdRevision, err = strconv.Atoi(versionTokens[6])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}

	result.CoreRevision, err = strconv.Atoi(versionTokens[7])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.Core18Revision, err = strconv.Atoi(versionTokens[8])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.Core20Revision, err = strconv.Atoi(versionTokens[9])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.Core22Revision, err = strconv.Atoi(versionTokens[10])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.Core24Revision, err = strconv.Atoi(versionTokens[11])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}

	result.M2cpGatewayRevision, err = strconv.Atoi(versionTokens[12])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.M2cpLogStatRevision, err = strconv.Atoi(versionTokens[13])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	result.M2cpMessageHubRevision, err = strconv.Atoi(versionTokens[14])
	if err != nil {
		return nil, fmt.Errorf("invalid os-version string: %s", err)
	}
	return &result, nil
}

func osVersion1FromString(short string) (*osVersionV1, error) {
	versionTokens := strings.Split(short, "-")
	if len(versionTokens) != 15 {
		return nil, fmt.Errorf("invalid os-version string")
	}

	switch versionTokens[0] {
	case "1":
		return parseVersion1(versionTokens)
	default:
		return nil, fmt.Errorf("invalid os-version string")
	}
}

func (dov *osVersionV1) Yaml() string {
	yamlBytes, err := yaml.Marshal(dov)
	if err != nil {
		return "error"
	}
	return string(yamlBytes)
}

func (dov *osVersionV1) Encode() (string, error) {
	snapstoreUrlCode, err := getValueForKey(storeCodesV1, dov.SnapstoreUrl)
	if err != nil {
		return "", fmt.Errorf("could not retrieve snapstore URL code: %s", err)
	}
	boardCode, err := getValueForKey(boardCodesV1, dov.Board)
	if err != nil {
		return "", fmt.Errorf("could not retrieve board code: %s", err)
	}
	tenantCode, err := getValueForKey(tenantCodesV1, dov.TenantAlias)
	if err != nil {
		return "", fmt.Errorf("could not retrieve tenant code: %s", err)
	}

	result := fmt.Sprintf("%d-%s-%s-%s-%d-%d-%d-%d-%d-%d-%d-%d-%d-%d-%d",
		dov.Version,
		snapstoreUrlCode, boardCode, tenantCode,
		dov.Kernel.Revision, dov.Gadget.Revision, dov.SnapdRevision,
		dov.CoreRevision, dov.Core18Revision, dov.Core20Revision, dov.Core22Revision, dov.Core24Revision,
		dov.M2cpGatewayRevision, dov.M2cpLogStatRevision, dov.M2cpMessageHubRevision)

	return result, nil
}
