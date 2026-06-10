package env

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type JsonWebToken struct {
	TenantId       string
	ExpirationTime float64
}

func NewJsonWebToken(jwt string) (*JsonWebToken, error) {
	if jwt == "" {
		return nil, nil
	}
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("could not split JWT")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode JWT")
	}

	result := JsonWebToken{}
	var claims map[string]interface{}
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshall JWT")
	}

	var ok bool
	result.TenantId, ok = claims["tenant_id"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find TenantId")
	}

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
