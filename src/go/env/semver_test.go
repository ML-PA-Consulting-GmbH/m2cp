package env

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestSemanticVersion_NewSemanticVersion_SimpleInput(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	str := "1.2.3"
	semver, err := NewSemanticVersion(str)
	assert.NoError(t, err)
	assert.NotNil(t, semver)

	assert.Equal(t, SemanticVersion{1, 2, 3, "", ""}, *semver)
	assert.True(t, semver.IsValid())
}

func TestSemanticVersion_NewSemanticVersion_ComplexInput(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	str := "1.2.3-beta+build.1-rc.10000aaa-kk-0.1"
	semver, err := NewSemanticVersion(str)
	assert.NoError(t, err)
	assert.NotNil(t, semver)

	assert.Equal(t, SemanticVersion{1, 2, 3, "beta", "build.1-rc.10000aaa-kk-0.1"}, *semver)
	assert.True(t, semver.IsValid())
}

func TestSemanticVersion_IsValid_ComplexInput(t *testing.T) {
	defer goleak.VerifyNone(t)
	semver := SemanticVersion{1, 2, 3, "beta", "build.1-rc.10000aaa-kk-0.1"}
	assert.True(t, semver.IsValid())
}

func TestSemanticVersion_String(t *testing.T) {
	defer goleak.VerifyNone(t)

	semver := SemanticVersion{1, 2, 3, "abc", "def"}
	str := semver.String()
	assert.Equal(t, "1.2.3-abc+def", str)
}

func TestSemanticVersion_NewSemanticVersion_ReadFromJSON(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	var semver SemanticVersion
	err = json.Unmarshal([]byte(`{
"major": 1,
"minor": 2,
"patch": 3,
"pre-release": "abc",
"build-metadata": "def"
}`), &semver)

	assert.NoError(t, err)
	assert.NotNil(t, semver)

	assert.Equal(t, SemanticVersion{1, 2, 3, "abc", "def"}, semver)
	assert.True(t, semver.IsValid())
}

func TestSemanticVersion_NewSemanticVersion_ReadFromYAML(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	var semver SemanticVersion
	err = yaml.Unmarshal([]byte(`{
"major": 1,
"minor": 2,
"patch": 3,
"pre-release": "abc",
"build-metadata": "def"
}`), &semver)

	assert.NoError(t, err)
	assert.NotNil(t, semver)

	assert.Equal(t, SemanticVersion{1, 2, 3, "abc", "def"}, semver)
	assert.True(t, semver.IsValid())
}
