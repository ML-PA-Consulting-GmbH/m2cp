package user

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd"
)

var UserCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"u"},
	Short:   "Manage the user session",
}

func init() {
	cmd.RootCmd.AddCommand(UserCmd)
}
