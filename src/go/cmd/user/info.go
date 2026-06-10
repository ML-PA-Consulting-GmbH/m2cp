package user

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/format"
	"m2cpcli/structs"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [userId | userEmail]",
	Short: "info about a user - by default the current one",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  runInfoCmd,
}

func init() {
	UserCmd.AddCommand(infoCmd)
}

func runInfoCmd(cmd *cobra.Command, args []string) (err error) {

	var user *structs.User
	if len(args) < 1 {
		user, err = backend.GetCurrentUser(cmd.Context())
		if err != nil {
			return err
		}
	} else if user, _ = backend.GetUserById(cmd.Context(), args[0]); user == nil {
		user, _ = backend.GetUserByEmail(cmd.Context(), args[0])
	}
	if user == nil {
		return fmt.Errorf("user '%s' not found", args[0])
	}

	return format.PrintFormattedOutput(cmd, user, nil)
}
