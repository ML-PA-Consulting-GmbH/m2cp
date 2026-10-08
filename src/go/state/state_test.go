package state

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The state file holds a bearer JWT; SecurePermissions must tighten a
// pre-existing world-readable file to owner-only, and must leave the
// containing directory untouched (--state may point at a repository root or $HOME).
func TestSecurePermissions_TightensExistingFileOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping POSIX permission tests on Windows")
	}

	dir := t.TempDir()
	sub := filepath.Join(dir, "m2cp")
	assert.NoError(t, os.Mkdir(sub, 0o755))
	assert.NoError(t, os.Chmod(sub, 0o755)) // Mkdir's mode is subject to umask
	statePath := filepath.Join(sub, "state.json")
	assert.NoError(t, os.WriteFile(statePath, []byte("{}"), 0o644))

	SecurePermissions(statePath)

	fi, err := os.Stat(statePath)
	assert.NoError(t, err)
	assert.Equal(t, FileMode, fi.Mode().Perm(), "state file must be 0600")
	assert.Equal(t, os.FileMode(0), fi.Mode().Perm()&0o077, "no group/other access")

	di, err := os.Stat(sub)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), di.Mode().Perm(), "containing dir must be left as-is")
}

// A missing path (first run, before the file exists) must be a harmless no-op.
func TestSecurePermissions_MissingPathIsNoop(t *testing.T) {
	assert.NotPanics(t, func() {
		SecurePermissions(filepath.Join(t.TempDir(), "does-not-exist", "state.json"))
	})
	assert.NoFileExistsf(t, filepath.Join(t.TempDir(), "does-not-exist", "state.json"), "state file should not exist")
	assert.NotPanics(t, func() {
		SecurePermissions("")
	})
}
