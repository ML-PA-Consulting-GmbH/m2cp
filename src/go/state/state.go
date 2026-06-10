package state

import (
	"m2cpcli/tools"
	"os"
)

const DefaultStatePath = "~/.m2cp/state.json"

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
