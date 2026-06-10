package tools

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
)

func TestAbspathWithEmptyString(t *testing.T) {
	var err error
	defer goleak.VerifyNone(t)

	actual, err := Abspath("")
	fmt.Println(actual)
	assert.NoError(t, err)
	// assert.True(t, strings.HasSuffix(actual, "/Tools/m2cp/tools"))
	// TODO: different path when running by script: git/M2CP/main/Tools/m2cp/bin/amd64/tests
}
