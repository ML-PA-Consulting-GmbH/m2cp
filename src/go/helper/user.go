package helper

import (
	"github.com/spf13/viper"
	"m2cpcli/env"
)

func UserIsLoggedIn() bool {
	jwtString := viper.GetString("jwt")
	jwt, err := env.NewJsonWebToken(jwtString)
	if err != nil {
		return false
	}
	if jwt.IsValid() {
		return true
	}
	return false
}
