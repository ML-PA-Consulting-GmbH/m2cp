package main

import (
	"m2cpcli/cmd"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// Visible sibling commands must not share a Short description, otherwise the
// help output lists two commands that look like they do the same thing.
// Hidden backward-compatibility aliases (e.g. fleet -> dgroup) are exempt.
func TestCommandTree_UniqueShortDescriptions(t *testing.T) {
	var walk func(parent *cobra.Command)
	walk = func(parent *cobra.Command) {
		seen := map[string]string{}
		for _, c := range parent.Commands() {
			if c.Short != "" && !c.Hidden {
				if other, ok := seen[c.Short]; ok {
					assert.Failf(t, "duplicate Short description",
						"%q: %q and %q share %q", parent.CommandPath(), other, c.Name(), c.Short)
				}
				seen[c.Short] = c.Name()
			}
			walk(c)
		}
	}
	walk(cmd.RootCmd)
}
