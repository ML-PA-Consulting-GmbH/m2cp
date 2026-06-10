package snapd

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"time"
)

type BaseAssertion struct {
	Type           string    `yaml:"type"`
	Revision       int       `yaml:"revision"`
	AuthorityId    string    `yaml:"authority-id"`
	Timestamp      time.Time `yaml:"timestamp"`
	SignKeySha3384 string    `yaml:"sign-key-sha3-384"`
	Signature      []byte    `yaml:"-"`
}

type ModelAssertion struct {
	BaseAssertion
	Series       int                  `yaml:"series"`
	BrandId      string               `yaml:"brand-id"`
	Model        string               `yaml:"model"`
	Architecture string               `yaml:"architecture"`
	Classic      bool                 `yaml:"classic"`
	Snaps        []ModelAssertionSnap `yaml:"snaps"`
}

type ModelAssertionSnap struct {
	DefaultChannel string `yaml:"default-channel"`
	Id             string `yaml:"id"`
	Name           string `yaml:"name"`
	Type           string `yaml:"type"`
}

func parseAssertion(assertionBytes []byte, assertionParsed interface{}) error {
	header, body, ok := bytes.Cut(assertionBytes, []byte("\n\n"))
	if !ok {
		return fmt.Errorf("could not split assertion")
	}

	err := yaml.Unmarshal(header, assertionParsed)
	if err != nil {
		return fmt.Errorf("could not parse assertion")
	}

	// Use type assertion to set the Signature field
	switch v := assertionParsed.(type) {
	case *BaseAssertion:
		v.Signature = body
	case *ModelAssertion:
		v.BaseAssertion.Signature = body
	default:
		return fmt.Errorf("unsupported struct provided as receiver for parsed data")
	}

	return nil
}

func ParseModelAssertion(assertionBytes []byte) (*ModelAssertion, error) {
	modelAssertion := ModelAssertion{}
	if err := parseAssertion(assertionBytes, &modelAssertion); err != nil {
		return nil, fmt.Errorf("could not parse model assertion: %s", err)
	}
	return &modelAssertion, nil
}
