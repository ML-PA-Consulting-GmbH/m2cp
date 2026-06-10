package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Store struct {
	Alias       string
	URL         string
	DockerImage string
}

func DockerImage(containerRegistryUrl, storeAlias string) string {
	return fmt.Sprintf("%s/m2cp-virtual-device-store-%s", containerRegistryUrl, storeAlias)
}

func StoreToAlias(url string) string {
	switch url {
	case "http://127.0.0.1:46900/graphql":
		return "local"
	case "http://localhost:46900/graphql":
		return "local"
	}

	sha := sha256URL(url)
	switch sha {
	case "a7d854c5b2de5a6cb713d4dc7fa9c52855d1c0095e67f2363e857844c3e68d9c":
		return "mlpa-dev"
	case "af19337882404d07b32282fda54bc6c2d7d411d3d08a7210bce202fecfb9d9d1":
		return "knorr-prod"
	case "28de3fcf10c442e32cdbdde55e1d827481bd1264e521aedde05cd3d7fcaeeb57":
		return "knorr-vtg-dev"
	}

	return "unknown"
}

func sha256URL(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

type StoreUrlAliasDefinition struct {
	data map[string]string
}

func GetAliases() (*StoreUrlAliasDefinition, error) {
	filepath, err := DefaultStoreUrlAliasDefinitionFilepath()
	if err != nil {
		return nil, err
	}
	aliasDefinition := NewStoreUrlAliasDefinition()
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		err = CreateFileIfNotExisting(filepath)
		if err != nil {
			return nil, err
		}
		err := aliasDefinition.Write(filepath)
		if err != nil {
			return nil, err
		}
	} else {
		err = aliasDefinition.ReadFromJson(filepath)
		if err != nil {
			return nil, err
		}
	}
	return aliasDefinition, nil
}

func NewStoreUrlAliasDefinition() *StoreUrlAliasDefinition {
	var result StoreUrlAliasDefinition
	result.data = map[string]string{}
	return &result
}

// ReadFromJson reads a dictionary from `filepath`. The structure of the JSON is simply `{"alias": "url", ...}`.
func (s *StoreUrlAliasDefinition) ReadFromJson(filepath string) error {
	jsonFile, err := os.OpenFile(filepath, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %s", err)
	}
	defer jsonFile.Close()

	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("could not read file %s: %s", filepath, err)
	}

	if len(jsonBytes) != 0 {
		err = json.Unmarshal(jsonBytes, &s.data)
		if err != nil {
			return fmt.Errorf("could not parse JSON from %s: %s", filepath, err)
		}
	}
	return nil
}

func StartsWithValidProtocol(url string) bool {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return true
	}
	return false
}

func (s *StoreUrlAliasDefinition) AddOrOverwrite(alias string, url string) error {
	if StartsWithValidProtocol(alias) {
		return fmt.Errorf("alias must not start with a valid protocol")
	}
	if !StartsWithValidProtocol(url) {
		return fmt.Errorf("URL must start with a valid protocol")
	}

	s.data[alias] = url
	return nil
}

func (s *StoreUrlAliasDefinition) GetUrl(alias string) (string, error) {
	if StartsWithValidProtocol(alias) {
		return "", fmt.Errorf("alias must not start with a valid protocol")
	}
	url, found := s.data[alias]
	if !found {
		return "", fmt.Errorf("could not retrieve store URL for alias \"%s\"", alias)
	}
	return url, nil
}

func (s *StoreUrlAliasDefinition) GetAlias(url string) string {
	for alias, storeUrl := range s.data {
		if storeUrl == url {
			return alias
		}
	}
	return ""
}

func (s *StoreUrlAliasDefinition) Write(filepath string) error {
	jsonBytes, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize JSON: %s", err)
	}

	jsonFile, err := os.OpenFile(filepath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %s", err)
	}
	defer jsonFile.Close()

	_, err = jsonFile.Write(jsonBytes)
	if err != nil {
		return fmt.Errorf("error writing file: %s", err)
	}
	return nil
}

func CreateFileIfNotExisting(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening file: %s", err)
		}
		err = file.Close()
		if err != nil {
			return fmt.Errorf("error closing file: %s", err)
		}
	}
	return nil
}

func DefaultStoreUrlAliasDefinitionFilepath() (string, error) {
	const filebase = "store-url-aliases"

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not retrieve home directory: %s", err)
	}
	filepath := path.Join(home, ".m2cp", fmt.Sprintf("%s.json", filebase))
	return filepath, nil
}

// SanitizeStoreUrl changes the given URL. Whitespace is stripped. If the url starts with a protocol, then a single
// trailing slash is removed.
func SanitizeStoreUrl(url string) (string, error) {
	url = strings.Trim(url, "\t\n\v\f\r ")
	if StartsWithValidProtocol(url) {
		if strings.HasSuffix(url, "/") {
			url = url[:len(url)-1]
		}
		// Note, according to https://graphql.org/learn/serving-over-http/#uris-routes, the single GraphQL endpoint
		// is "usually" called `/graphql`, but it is not a must! Automatically sanitizing this suffix means, we have to
		// recompile, once somebody decides to use a different URL. And we risk sloppy documentation, were the suffix is
		// omitted.
		const graphqlSuffix = "/graphql"
		if !strings.HasSuffix(url, graphqlSuffix) {
			url += graphqlSuffix
		}
	}
	return url, nil
}
