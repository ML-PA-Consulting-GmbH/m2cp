package tenant

import (
	"github.com/spf13/cobra"
	"m2cpcli/cmd/user"
)

var tenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage the user tenant",
}

func init() {
	user.UserCmd.AddCommand(tenantCmd)
}
