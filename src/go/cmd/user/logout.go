package user

import (
	"fmt"
	"m2cpcli/backend"
	"m2cpcli/env"
	"m2cpcli/format"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout current user",
	Args:  cobra.ExactArgs(0),
	RunE:  runLogoutCmd,
}

func init() {
	UserCmd.AddCommand(logoutCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// UserCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// UserCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runLogoutCmd(cmd *cobra.Command, args []string) error {
	var LogoutResult struct {
		Message string `json:"message"`
	}

	jwtString := viper.GetString("jwt")
	jwt, err := env.NewJsonWebToken(jwtString)
	if err != nil {
		return err
	}
	if jwt.IsValid() {
		success, err := backend.UserLogout(cmd.Context())
		if err != nil {
			return err
		}

		if success {
			LogoutResult.Message = "logged out successfully"
		} else {
			LogoutResult.Message = "failed to log out"
		}

		err = env.InvalidateSessionJwt()
		if err != nil {
			return fmt.Errorf("failed to invalidate session jwt: %v", err)
		}
	} else {
		LogoutResult.Message = "logged out already"
	}

	return format.PrintFormattedOutput(cmd, LogoutResult, nil)
}
