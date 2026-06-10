package roles

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/user"
)

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage the user roles",
}

func init() {
	user.UserCmd.AddCommand(rolesCmd)
}
