package state

import (
	"m2cpcli/tools"
	"os"
)

const DefaultStatePath = "~/.m2cp/state.json"

// State file/directory permission modes.
// The state file holds a bearer JWT, so it must not be readable by group/other.
// DirMode applies only to directories the CLI creates itself (see createFile);
// an existing directory is never re-permissioned.
const (
	FileMode = os.FileMode(0o600)
	DirMode  = os.FileMode(0o700)
)

var ConfigStateFileName = ""

// FixLegacyStateFile checks if the legacy state file exists and moves it to the new name if necessary.
func FixLegacyStateFile() {
	legacyPath, _ := tools.Abspath("~/.m2cp/state2.json")
	newPath, _ := tools.Abspath(DefaultStatePath)
	if _, err := os.Stat(legacyPath); err == nil {
		if _, err := os.Stat(newPath); err == nil {
			_ = os.Remove(legacyPath)
			return
		}
		_ = os.Rename(legacyPath, DefaultStatePath)
	}
}

// SecurePermissions limits the state file to owner-only access (0600).
// The file may have been created by an older CLI (or migrated from state2.json) with the
// default world-readable 0644 mode, and viper's WriteConfig only applies its
// configured permissions when it first creates the file — never to an existing one.
// The containing directory is deliberately left alone: --state may point anywhere
// (a repository root, $HOME), and 0600 protects the file regardless of the directory mode.
// Missing paths are ignored; chmod errors are non-fatal (e.g. a read-only filesystem) so this never blocks startup.
// Note that this is a no-op on Windows, where the file mode bits are not meaningful.
func SecurePermissions(path string) {
	if path == "" {
		return
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		_ = os.Chmod(path, FileMode)
	}
}
