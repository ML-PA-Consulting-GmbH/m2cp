package env

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type JsonWebToken struct {
	ExpirationTime float64
}

// DecodeJwtClaims decodes the payload segment of a JWT into a claims map.
func DecodeJwtClaims(jwt string) (map[string]interface{}, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("could not split JWT")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode JWT")
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("could not unmarshall JWT")
	}
	return claims, nil
}

func NewJsonWebToken(jwt string) (*JsonWebToken, error) {
	if jwt == "" {
		return nil, nil
	}

	claims, err := DecodeJwtClaims(jwt)
	if err != nil {
		return nil, err
	}

	result := JsonWebToken{}

	var ok bool
	result.ExpirationTime, ok = claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("could not find ExpirationTime")
	}
	return &result, nil
}

func (jwt *JsonWebToken) IsValid() bool {
	if jwt == nil {
		return false
	}
	expiresTime := time.Unix(int64(jwt.ExpirationTime), 0)
	if expiresTime.Before(time.Now()) {
		return false
	}
	return true
}
