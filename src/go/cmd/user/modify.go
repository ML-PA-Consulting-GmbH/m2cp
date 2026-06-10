package user

import (
	"fmt"
	"github.com/spf13/cobra"
	"m2cpcli/format"
	gql "m2cpcli/graphql"
	"os"
	"strings"
)

var modifyCmd = &cobra.Command{
	Use:   "modify",
	Short: "Modify user details",
	Args:  cobra.ExactArgs(0),
	RunE:  runModifyCmd,
}

func init() {
	UserCmd.AddCommand(modifyCmd)
	modifyCmd.Flags().String("ssh-key", "", "new SSH public key of the user")
}

func sshKeySuperficiallyLooksValid(sshKey string) bool {
	if len(sshKey) > 4 && strings.HasPrefix(sshKey, "ssh-") {
		return true
	}
	return false
}

func runModifyCmd(cmd *cobra.Command, args []string) error {
	var sshKey string
	var err error

	if cmd.Flags().Changed("ssh-key") {
		sshKey, err = cmd.Flags().GetString("ssh-key")
		if err != nil {
			return err
		}
		// check, if sshKey is a path to a file - and if so, read the content
		if strings.HasPrefix(sshKey, "ssh-rsa ") {
			// sshKey is a value - leave it as is
			sshKey = strings.TrimSpace(sshKey)
		} else {
			// let's assume it's a path to a file - and see if the file exists
			if fileInfo, err := os.Stat(sshKey); err == nil && !fileInfo.IsDir() {
				content, err := os.ReadFile(sshKey)
				if err != nil {
					return fmt.Errorf("failed to read ssh-key file: %v", err)
				}
				sshKey = strings.TrimSpace(string(content))
			}
		}

		if !sshKeySuperficiallyLooksValid(sshKey) {
			return fmt.Errorf("ssh-key value is invalid")
		}
	} else {
		return fmt.Errorf("ssh-key flag is required")
	}

	// Get user, we require the id for any modification
	user, err := gql.UserInfo(cmd.Context())
	if err != nil {
		return err
	}

	err = gql.UserModify(cmd.Context(), user.Id, sshKey)
	if err != nil {
		return err
	}

	var UserModifyResult struct {
		Message string `json:"message"`
	}
	UserModifyResult.Message = "modified successfully"

	return format.PrintFormattedOutput(cmd, UserModifyResult, nil)
}
