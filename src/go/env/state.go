package env

import (
	"encoding/json"
	"errors"
	"fmt"
	"m2cpcli/tools"
	"os"
	"path/filepath"
	"time"
)

type PersistedData struct {
	Url            string `json:"url"`
	JSONWebToken   string `json:"jwt"`
	UserEmail      string `json:"user"`
	PrivateKeyPath string `json:"key"`
	DeviceHub      string `json:"device_hub"`
	Tenant         struct {
		Name  string `json:"name,omitempty"`
		Alias string `json:"alias,omitempty"`
		Id    string `json:"id"` // tenantId comes from the JWT
	} `json:"tenant"`
	Docker struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"docker"`
}

type PersistedState struct {
	filepath string
	Data     *PersistedData
}

func NewPersistedState(filepath string) (*PersistedState, error) {
	var err error
	state := PersistedState{}
	state.filepath, err = tools.Abspath(filepath)
	if err != nil {
		return nil, err
	}
	state.Data = &PersistedData{}

	if _, err := os.Stat(state.filepath); errors.Is(err, os.ErrNotExist) {
		//fmt.Printf("creating a new '%s'\n", state.filepath) // TODO: debug output
		err = state.Save()
		if err != nil {
			return nil, err
		}
	} else {
		//fmt.Printf("'%s' exists\n", state.filepath) // TODO: debug output
		state.Load()
	}

	return &state, nil
}

func (s *PersistedState) Load() error {
	var err error
	stateJson, err := tools.ReadLocalFile(s.filepath)
	if err != nil {
		return fmt.Errorf("could not read file '%s'", s.filepath)
	}
	var newState PersistedData
	err = json.Unmarshal(stateJson, &newState)
	if err != nil {
		return fmt.Errorf("could not parse file '%s'", s.filepath)
	}
	*(s.Data) = newState
	return nil
}

func (s *PersistedState) Save() error {
	dir := filepath.Dir(s.filepath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(s.filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	content, err := json.Marshal(s.Data)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	return err
}

func (s *PersistedState) Login(url, userEmail, privateKeyPath, jwt string) error {
	var err error
	s.Data.UserEmail = userEmail
	s.Data.JSONWebToken = jwt
	s.Data.Url = url
	s.Data.PrivateKeyPath, err = tools.Abspath(privateKeyPath)
	if err != nil {
		return fmt.Errorf("could not get private key path")
	}

	var jwtObject *JsonWebToken
	jwtObject, err = NewJsonWebToken(s.Data.JSONWebToken)
	if err != nil {
		return err
	}
	s.Data.Tenant.Id = jwtObject.TenantId

	err = s.Save()
	return err
}

func (s *PersistedState) Logout() {
	s.Data.JSONWebToken = ""
	err := s.Save()
	if err != nil {
		fmt.Println("warning: could not persist state after logout")
	}
}

func (s *PersistedState) IsAuthenticated() bool {
	if s.Data == nil || s.Data.JSONWebToken == "" {
		return false
	}

	jwt, err := NewJsonWebToken(s.Data.JSONWebToken)
	if err != nil {
		return false
	}

	expiresTime := time.Unix(int64(jwt.ExpirationTime), 0)
	if expiresTime.Before(time.Now()) {
		return false
	}

	return true
}

func (s *PersistedState) GetBackendUrl() string {
	return s.Data.Url
}

func (s *PersistedState) GetBackendAuthorization() string {
	return fmt.Sprintf("Bearer %s", s.Data.JSONWebToken)
}

func (s *PersistedState) GetAuthHeader() *map[string]string {
	panic("deprecated")
}
