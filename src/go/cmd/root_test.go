package cmd

import (
	"m2cpcli/state"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createFile must create the state file (which holds a bearer JWT) owner-only,
// not with os.Create's world-readable 0644, and lock down the directory too.
func TestCreateFile_UsesOwnerOnlyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping POSIX permission tests on Windows")
	}

	// Nested path so MkdirAll has a directory to create.
	statePath := filepath.Join(t.TempDir(), "m2cp", "state.json")

	createFile(statePath)

	fi, err := os.Stat(statePath)
	assert.NoError(t, err)
	assert.Equal(t, state.FileMode, fi.Mode().Perm(), "state file must be 0600")

	di, err := os.Stat(filepath.Dir(statePath))
	assert.NoError(t, err)
	assert.Equal(t, state.DirMode, di.Mode().Perm(), "state dir must be 0700")
}

func TestInitConfiguration(t *testing.T) {
	state.ConfigStateFileName = filepath.Join(t.TempDir(), "m2cp", "state.json")
	err := initConfiguration()
	assert.NoError(t, err)

	// Check that the state file was created
	_, err = os.Stat(state.ConfigStateFileName)
	assert.NoError(t, err)
}
