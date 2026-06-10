package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const testContainerRegistryUri string = "acrtest.azurecr.io"

func TestDockerImage_IntegrationTests(t *testing.T) {
	url := testContainerRegistryUri
	storeAlias := "integration-tests"
	actual := DockerImage(url, storeAlias)

	assert.Equal(t, "acrtest.azurecr.io/m2cp-virtual-device-store-integration-tests", actual)
}

func TestDockerImage_Local(t *testing.T) {
	url := testContainerRegistryUri
	storeAlias := "local"
	actual := DockerImage(url, storeAlias)

	assert.Equal(t, "acrtest.azurecr.io/m2cp-virtual-device-store-local", actual)
}

func TestDockerImage_MlpaDev(t *testing.T) {
	url := testContainerRegistryUri
	storeAlias := "mlpa-dev"
	actual := DockerImage(url, storeAlias)

	assert.Equal(t, "acrtest.azurecr.io/m2cp-virtual-device-store-mlpa-dev", actual)
}

func TestDockerImage_MlpaTest(t *testing.T) {
	url := testContainerRegistryUri
	storeAlias := "mlpa-test"
	actual := DockerImage(url, storeAlias)

	assert.Equal(t, "acrtest.azurecr.io/m2cp-virtual-device-store-mlpa-test", actual)
}

func TestSanitizeStoreUrl_AliasWithWhitespace(t *testing.T) {
	url := " alias\t"
	sanitizedUrl, err := SanitizeStoreUrl(url)
	assert.NoError(t, err)
	assert.Equal(t, "alias", sanitizedUrl)
}

func TestSanitizeStoreUrl_HttpUrlWithWhitespaceWithoutGraphqlSuffix(t *testing.T) {
	url := "\nhttp://example.com  "
	sanitizedUrl, err := SanitizeStoreUrl(url)
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/graphql", sanitizedUrl)
}

func TestSanitizeStoreUrl_SlashGraphqlSlashAsAlias(t *testing.T) {
	url := "/graphql/"
	sanitizedUrl, err := SanitizeStoreUrl(url)
	assert.NoError(t, err)
	assert.Equal(t, "/graphql/", sanitizedUrl)
}

func TestSanitizeStoreUrl_HttpsAsAlias(t *testing.T) {
	url := "https"
	sanitizedUrl, err := SanitizeStoreUrl(url)
	assert.NoError(t, err)
	assert.Equal(t, "https", sanitizedUrl)
}

func TestSanitizeStoreUrl_UrlWithGraphqlSuffixIsUnchanged(t *testing.T) {
	url := "https://example.com/graphql"
	sanitizedUrl, err := SanitizeStoreUrl(url)
	assert.NoError(t, err)
	assert.Equal(t, url, sanitizedUrl)
}
