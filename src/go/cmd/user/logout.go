package user

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"m2cpcli/env"
	"m2cpcli/format"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout current developer",
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
	jwtString := viper.GetString("jwt")

	var LogoutResult struct {
		Message string `json:"message"`
	}

	jwt, err := env.NewJsonWebToken(jwtString)
	if err != nil {
		return err
	}
	if jwt.IsValid() {
		// TODO: server side invalidation was removed
		//err = auth.InvalidateJSONWebToken(cmd.Context(), url, jwtString)
		//if err != nil {
		//	return err
		//}
	}
	err = env.InvalidateSessionJwt()
	if err != nil {
		return err
	}
	// either way the same message
	LogoutResult.Message = "logged out successfully"

	return format.PrintFormattedOutput(cmd, LogoutResult, nil)
}
