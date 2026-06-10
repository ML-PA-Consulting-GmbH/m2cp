package env

import "fmt"

type GitMetadata struct {
	Commit string `json:"commit"`
	// Branch string `json:"branch"`
	// UserEmail string `json:"user.email"` // key as in git config --list
	// Host?
	// User?
	// Git Tag?
}

func (g *GitMetadata) String() string {
	str := fmt.Sprintf("git.%s", g.Commit)
	return str
}
