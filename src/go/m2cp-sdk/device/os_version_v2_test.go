package device

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"m2cp/contextplus"
	"os"
	"testing"
)

func TestGetOsVersion2(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping in CI environment because snapd/virtual device is not available")
		return
	}
	ctp := contextplus.NewContextPlus()
	v2, err := getOsVersion2(ctp)
	assert.NoError(t, err)
	assert.NotEmpty(t, v2)
	fmt.Println("os-version-v2: " + v2)

}
