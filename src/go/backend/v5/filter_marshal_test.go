package v5

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComparableNullableOfDateTimeOperationFilterInput_OmitsUnsetFields(t *testing.T) {
	date := "2026-06-18"
	filter := ComparableNullableOfDateTimeOperationFilterInput{Gte: &date}

	body, err := json.Marshal(filter)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"gte":"2026-06-18"}`, string(body))
}

func TestStringOperationFilterInput_OmitsUnsetFields(t *testing.T) {
	value := "foo"
	filter := StringOperationFilterInput{Eq: &value}

	body, err := json.Marshal(filter)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"eq":"foo"}`, string(body))
}

func TestArchitectureOperationFilterInput_OmitsUnsetFields(t *testing.T) {
	arch := Architecture("ARM64")
	filter := ArchitectureOperationFilterInput{Eq: &arch}

	body, err := json.Marshal(filter)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"eq":"ARM64"}`, string(body))
}

func TestComparableGuidOperationFilterInput_OmitsUnsetFields(t *testing.T) {
	id := "8f4d8c1e-7d91-4bd4-15f0-08ded81ac049"
	filter := ComparableGuidOperationFilterInput{Eq: &id}

	body, err := json.Marshal(filter)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"eq":"8f4d8c1e-7d91-4bd4-15f0-08ded81ac049"}`, string(body))
}
