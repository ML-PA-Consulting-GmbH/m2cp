package user

import (
	"m2cpcli/env"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var jwtCmd = &cobra.Command{
	Use:   "jwt",
	Short: "print current jwt token of the active session",
	Args:  cobra.ExactArgs(0),
	RunE:  runJwtCmd,
}

func init() {
	UserCmd.AddCommand(jwtCmd)
}

func runJwtCmd(cmd *cobra.Command, args []string) error {
	jwtString := viper.GetString("jwt")
	jwt, err := env.NewJsonWebToken(jwtString)
	if err != nil {
		return err
	}
	if !jwt.IsValid() {
		err = env.InvalidateSessionJwt()
		if err != nil {
			return err
		}
	}
	println(jwtString)
	return nil
}
