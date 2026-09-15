package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAuthenticationMethod_M2M(t *testing.T) {
	m, err := ParseAuthenticationMethod("m2m")
	assert.NoError(t, err)
	assert.Equal(t, M2MAuthentication, m)

	// case-insensitive, matching the existing ssh/browser behavior
	m, err = ParseAuthenticationMethod("M2M")
	assert.NoError(t, err)
	assert.Equal(t, M2MAuthentication, m)
}

func TestAuthenticationMethod_String_M2M(t *testing.T) {
	assert.Equal(t, "m2m", M2MAuthentication.String())
}

func TestListingOfKnownAuthenticationMethods_IncludesAll(t *testing.T) {
	listing := ListingOfKnownAuthenticationMethods()
	assert.Contains(t, listing, `"m2m"`)
	assert.Contains(t, listing, `"ssh"`)
	assert.Contains(t, listing, `"browser"`)
}

func TestParseAuthenticationMethod_Invalid(t *testing.T) {
	// the old method name "client" was renamed to "m2m" and must now be rejected
	_, err := ParseAuthenticationMethod("client")
	assert.Error(t, err)
}
