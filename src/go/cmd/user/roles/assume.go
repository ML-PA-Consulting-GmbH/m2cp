package roles

import (
	"fmt"
	"github.com/spf13/cobra"
)

var assumeCmd = &cobra.Command{
	Use:   "assume <email>",
	Short: "Assume the roles assigned to another user (super admin only)",
	Long: `Assume the roles assigned to another user without logging in as that user. 
			This command is only available to super admins and does not grant access to the user's specific resources, only their roles.`,
	Args: cobra.ExactArgs(1),
	RunE: runAssumeCmd,
}

type assumeOutput struct {
	JwtToken string
}

func init() {
	rolesCmd.AddCommand(assumeCmd)
}

func runAssumeCmd(cmd *cobra.Command, args []string) error {
	/*
		email := args[0]
		if email == "" {
			return fmt.Errorf("email is required")
		}

		getUserIdResponse, err := backend.GetUserIdByEmail(cmd.Context(), email)
		if err != nil {
			return err
		}

		if getUserIdResponse.Users == nil || getUserIdResponse.Users.Items == nil || len(getUserIdResponse.Users.Items) == 0 {
			return fmt.Errorf("user not found")
		}

		userId := getUserIdResponse.Users.Items[0].Id

		loginAsResponse, err := backend.UserLoginAs(cmd.Context(), userId)
		if err != nil {
			return err
		}

		if loginAsResponse.UserLoginAs == nil || loginAsResponse.UserLoginAs.Token == "" {
			return fmt.Errorf("failed to login as user")
		}

		viper.Set("jwt", loginAsResponse.UserLoginAs.Token)
		err = viper.WriteConfig()

		if err != nil {
			return err
		}

		output := assumeOutput{
			JwtToken: loginAsResponse.UserLoginAs.Token,
		}

		err = format.PrintFormattedOutput(cmd, output, assumeOutputFormatter)
		if err != nil {
			return err
		}

		return nil

	*/
	return nil
}

func assumeOutputFormatter(_ assumeOutput) (string, error) {
	outputStr := ""

	outputStr += fmt.Sprintf("JWT Token successfully generated. \n* Please note that if you want to switch back to your original user, you need to logout and login again.")

	return outputStr, nil
}
