package tools

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
)

func TestIsValidSnapName_Positives(t *testing.T) {
	defer goleak.VerifyNone(t)

	validNames := []string{"a1243", "1a234", "123a4", "1234a", "snap-mi36citu"}
	for _, name := range validNames {
		assert.True(t, IsValidAppName(name))
	}
}

func TestIsValidSnapName_Negatives(t *testing.T) {
	defer goleak.VerifyNone(t)

	invalidNames := []string{"1243", "under_score", "über", "1234-", "-1234", "a-", "-a", "--", "ABC"}
	for _, name := range invalidNames {
		assert.False(t, IsValidAppName(name))
	}
}

func TestIsValidSnapVersion_Positives(t *testing.T) {
	defer goleak.VerifyNone(t)

	validVersions := []string{"1", "1.2", "1.2.3", "1.2.3.4", "1-beta", "1.0.0-alpha", "0-nothing", "0-007"}
	for _, name := range validVersions {
		assert.True(t, IsValidSnapVersion(name))
	}
}

func TestIsValidUuid_Positives(t *testing.T) {
	defer goleak.VerifyNone(t)

	validUuids := []string{
		"226df810-b146-459e-bc64-defdfc5e6062",
		"7275f4de-6136-4aa2-9fcb-67d0174b7c8b",
		"18271999-1e75-45d5-8677-4bb64c77eebd",
		"00000000-0000-0000-0000-000000000000",
		"00000000-0000-0000-0000-000000000001",
	}
	for _, name := range validUuids {
		assert.True(t, IsValidUuid(name))
	}
}

func TestIsValidUuid_Negatives(t *testing.T) {
	defer goleak.VerifyNone(t)

	invalidUuids := []string{
		"",
		"00000000-0000-0000-0000-00000000000",
		"000000000-000-0000-0000-000000000000",
		"00000000-0000-0000-0000-0000000000012",
		"7275f4de61364aa29fcb67d0174b7c8b",
	}
	for _, name := range invalidUuids {
		assert.False(t, IsValidUuid(name))
	}
}

func TestIsValidSnapId_Positives(t *testing.T) {
	defer goleak.VerifyNone(t)

	validUuids := []string{
		"00000000000000000000000000000000",
		"00000000000000000000000000000001",
		"226df810b146459ebc64defdfc5e6062",
		"7275f4de61364aa29fcb67d0174b7c8b",
		"182719991e7545d586774bb64c77eebd",
	}
	for _, name := range validUuids {
		assert.True(t, IsValidAppId(name))
	}
}

func TestIsValidSnapId_Negatives(t *testing.T) {
	defer goleak.VerifyNone(t)

	validUuids := []string{
		"",
		"226df810-b146-459e-bc64-defdfc5e6062",
		"7275f4de6136-4aa2-9fcb-67d0174b7c8b",
		"182719991e7545d5-8677-4bb64c77eebd",
		"000000000000000000000-000000000000",
		"00000000000000000000000000000000x",
		"0000000000000000000000000000002",
		"000000000000000000000000000003",
		"00000000000000000000000000000000x",
		"0000000000000000000000000000000g",
	}
	for _, name := range validUuids {
		assert.False(t, IsValidAppId(name))
	}
}

func TestSplitIntoEqualChunks_NormalCase(t *testing.T) {
	defer goleak.VerifyNone(t)

	encoded := []byte("badcaffee")
	chunks := SplitIntoEqualChunks(encoded, 4)
	assert.Equal(t, []byte("badc"), chunks[0])
	assert.Equal(t, []byte("affe"), chunks[1])
	assert.Equal(t, []byte("e"), chunks[2])
}

func TestSplitIntoEqualChunks_EmptyString(t *testing.T) {
	defer goleak.VerifyNone(t)

	encoded := []byte("")
	chunks := SplitIntoEqualChunks(encoded, 4)
	assert.Equal(t, []byte(""), chunks[0])
}

func TestSplitIntoEqualChunks_ZeroChunkSize(t *testing.T) {
	defer goleak.VerifyNone(t)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	encoded := []byte("panic")
	_ = SplitIntoEqualChunks(encoded, 0)
}
