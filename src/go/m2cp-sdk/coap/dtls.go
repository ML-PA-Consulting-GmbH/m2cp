package coap

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"m2cp"
	"os"
	"strings"

	"github.com/google/uuid"
	piondtls "github.com/pion/dtls/v3"
	"github.com/pion/logging"
)

const (
	// This will still be overshadowed of the log level set in the ContextPlus
	// However it still discerns between trace and debug if debug set is set in the ContextPlus
	defaultDTLSLogLevel = logging.LogLevelDebug
)

func NewDTLSConfigSingleKey(ownIdentity string, psk []byte) m2cp.DTLSConfig {

	return func(ctp m2cp.ContextPlus) (*piondtls.Config, m2cp.Identity) {
		return &piondtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				ctp.LogDebug("Peers identity hint: %s", string(hint))
				return psk, nil
			},
			PSKIdentityHint: []byte(ownIdentity),
			CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
			LoggerFactory:   NewPionLoggerFactory(ctp, defaultDTLSLogLevel),
		}, m2cp.Identity{Identity: "reachability-test", Key: psk}
	}

}

func NewDTLSConfigWithFunc(ownIdentity string, pskFunc func(identityHint string) ([]byte, error)) m2cp.DTLSConfig {
	// create an additional random identity to be used for reachability testing
	key := make([]byte, 16)
	//from the documentation of crypto/rand.Read: It never returns an error, and always fills b entirely.
	_, _ = rand.Read(key)
	reachID := m2cp.Identity{
		Identity: uuid.New().String(),
		Key:      key,
	}

	return func(ctp m2cp.ContextPlus) (*piondtls.Config, m2cp.Identity) {
		return &piondtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				ctp.LogDebug("Peers identity hint: %s", string(hint))
				if string(hint) == reachID.Identity {
					return reachID.Key, nil
				}
				return pskFunc(string(hint))
			},
			PSKIdentityHint: []byte(ownIdentity),
			CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
			LoggerFactory:   NewPionLoggerFactory(ctp, defaultDTLSLogLevel),
		}, reachID
	}
}

func NewDTLSConfigFromMap(ownIdentity string, identities map[string][]byte) m2cp.DTLSConfig {

	// create an additional random identity to be used for reachability testing
	key := make([]byte, 16)
	//from the documentation of crypto/rand.Read: It never returns an error, and always fills b entirely.
	_, _ = rand.Read(key)

	reachID := m2cp.Identity{
		Identity: uuid.New().String(),
		Key:      key,
	}
	identities[reachID.Identity] = reachID.Key

	return func(ctp m2cp.ContextPlus) (*piondtls.Config, m2cp.Identity) {
		return &piondtls.Config{
			PSK: func(hint []byte) ([]byte, error) {
				if psk, ok := identities[string(hint)]; ok {
					ctp.LogDebug("Peers identity hint: %s, found respective key", string(psk))
					return psk, nil
				}

				return nil, fmt.Errorf("unknown identity hint: %s", string(hint))
			},
			PSKIdentityHint: []byte(ownIdentity),
			CipherSuites:    []piondtls.CipherSuiteID{piondtls.TLS_PSK_WITH_AES_128_CCM_8},
			LoggerFactory:   NewPionLoggerFactory(ctp, defaultDTLSLogLevel),
		}, reachID
	}
}

// NewDTLSConfigFromFile creates a DTLSConfig by reading multiple PSK identities from a file.
// Each line should have the format  `identity: hexkey`
func NewDTLSConfigFromFile(ownIdentity string, pskFilePath string) (m2cp.DTLSConfig, error) {
	content, err := os.ReadFile(pskFilePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read PSK file: %w", err)
	}
	identities, err := ParsePSKString(string(content))
	if err != nil {
		return nil, fmt.Errorf("cannot parse PSK file: %w", err)
	}
	return NewDTLSConfigFromMap(ownIdentity, identities), nil
}

// ParsePSKString parses the contents of a PSK file
// Each line should have the format  `identity: hexkey`
// It returns a map of identity string and hexkey as bytes and an error if parsing fails.
func ParsePSKString(content string) (map[string][]byte, error) {
	identities := make(map[string][]byte)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line format: %s", line)
		}
		identity := strings.TrimSpace(parts[0])
		hexKey := strings.TrimSpace(parts[1])
		key, err := hex.DecodeString(hexKey)
		if err != nil {
			return nil, fmt.Errorf("invalid hex key for identity %s: %w", identity, err)
		}
		identities[identity] = key
	}
	return identities, nil
}
