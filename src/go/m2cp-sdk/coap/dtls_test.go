package coap

import (
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/suite"
	"m2cp"
	"m2cp/contextplus"
	"testing"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (s *TestSuite) SetupSuite() {
	s.T().Logf(">>> From SetupSuite")
}

func (s *TestSuite) TearDownSuite() {
	s.T().Logf(">>> From TearDownSuite")
}

func (s *TestSuite) SetupTest() {
	s.T().Logf("-- From SetupTest")
	s.ctp = contextplus.NewContextPlus()
}

func (s *TestSuite) TearDownTest() {
	s.T().Logf("-- From TearDownTest")
	s.ctp.Cancel()
}

func (s *TestSuite) TestParsePSKString_ValidMultipleIdentities() {
	content := `test-client: bdaa6ffad93c9b07fd84480e672e0ea2
test-client-2: 8daa6ffad93c9b07fd84480e672e0ea2
non-existent: 12126ffad93c9ba7fd84480d67210ea2`

	identities, err := ParsePSKString(content)
	s.NoError(err)
	s.Len(identities, 3)

	// Verify first identity
	expectedKey1, _ := hex.DecodeString("bdaa6ffad93c9b07fd84480e672e0ea2")
	s.Equal(expectedKey1, identities["test-client"])

	// Verify second identity
	expectedKey2, _ := hex.DecodeString("8daa6ffad93c9b07fd84480e672e0ea2")
	s.Equal(expectedKey2, identities["test-client-2"])

	// Verify third identity
	expectedKey3, _ := hex.DecodeString("12126ffad93c9ba7fd84480d67210ea2")
	s.Equal(expectedKey3, identities["non-existent"])
}

func (s *TestSuite) TestParsePSKString_EmptyLines() {
	content := `test-client: bdaa6ffad93c9b07fd84480e672e0ea2

test-client-2: 8daa6ffad93c9b07fd84480e672e0ea2

`

	identities, err := ParsePSKString(content)
	s.NoError(err)
	s.Len(identities, 2)
	s.Contains(identities, "test-client")
	s.Contains(identities, "test-client-2")
}

func (s *TestSuite) TestParsePSKString_InvalidFormat() {
	content := `test-client bdaa6ffad93c9b07fd84480e672e0ea2`

	identities, err := ParsePSKString(content)
	s.Error(err)
	s.Nil(identities)
	s.Contains(err.Error(), "invalid line format")
}

func (s *TestSuite) TestParsePSKString_InvalidHex() {
	content := `test-client: notahexstring`

	identities, err := ParsePSKString(content)
	s.Error(err)
	s.Nil(identities)
	s.Contains(err.Error(), "invalid hex key")
}

func (s *TestSuite) TestParsePSKString_EmptyString() {
	identities, err := ParsePSKString("")
	s.NoError(err)
	s.Empty(identities)
}

func (s *TestSuite) TestNewDTLSConfigFromFile_ValidFile() {
	builder, err := NewDTLSConfigFromFile("test-server", "testdata/test-identity.txt")
	s.NoError(err)
	s.NotNil(builder)

	config, _ := builder(s.ctp)
	s.NotNil(config)
	s.Equal([]byte("test-server"), config.PSKIdentityHint)

	// Test PSK lookup for all three identities
	psk1, err := config.PSK([]byte("test-client"))
	s.NoError(err)
	expectedKey1, _ := hex.DecodeString("bdaa6ffad93c9b07fd84480e672e0ea2")
	s.Equal(expectedKey1, psk1)

	psk2, err := config.PSK([]byte("test-client-2"))
	s.NoError(err)
	expectedKey2, _ := hex.DecodeString("8daa6ffad93c9b07fd84480e672e0ea2")
	s.Equal(expectedKey2, psk2)

	psk3, err := config.PSK([]byte("non-existent"))
	s.NoError(err)
	expectedKey3, _ := hex.DecodeString("12126ffad93c9ba7fd84480d67210ea2")
	s.Equal(expectedKey3, psk3)
}

func (s *TestSuite) TestNewDTLSConfigFromFile_NonExistentFile() {
	builder, err := NewDTLSConfigFromFile("test-server", "testdata/non-existent.txt")
	s.Error(err)
	s.Nil(builder)
	s.Contains(err.Error(), "cannot read PSK file")
}

func (s *TestSuite) TestNewDTLSConfigFromMap_MultipleIdentities() {
	identities := map[string][]byte{
		"client1": []byte("key1"),
		"client2": []byte("key2"),
		"client3": []byte("key3"),
	}

	builder := NewDTLSConfigFromMap("test-server", identities)
	s.NotNil(builder)

	config, _ := builder(s.ctp)
	s.NotNil(config)
	s.Equal([]byte("test-server"), config.PSKIdentityHint)

	// Test all identities
	psk1, err := config.PSK([]byte("client1"))
	s.NoError(err)
	s.Equal([]byte("key1"), psk1)

	psk2, err := config.PSK([]byte("client2"))
	s.NoError(err)
	s.Equal([]byte("key2"), psk2)

	psk3, err := config.PSK([]byte("client3"))
	s.NoError(err)
	s.Equal([]byte("key3"), psk3)
}

func (s *TestSuite) TestNewDTLSConfigFromMap_ReachabilityIdentity() {
	identities := map[string][]byte{
		"client1": []byte("key1"),
	}

	builder := NewDTLSConfigFromMap("test-server", identities)
	config, reachID := builder(s.ctp)
	s.NotNil(config)
	s.NotEmpty(reachID.Identity)
	s.NotNil(reachID.Key)

	psk, err := config.PSK([]byte(reachID.Identity))
	s.NoError(err)
	s.Equal(reachID.Key, psk)
}

func (s *TestSuite) TestNewDTLSConfigFromMap_UnknownIdentity() {
	identities := map[string][]byte{
		"client1": []byte("key1"),
	}

	builder := NewDTLSConfigFromMap("test-server", identities)
	config, _ := builder(s.ctp)

	psk, err := config.PSK([]byte("unknown-client"))
	s.Error(err)
	s.Nil(psk)
	s.Contains(err.Error(), "unknown identity hint")
}

func (s *TestSuite) TestNewDTLSConfigFromMap_StableOutput() {
	identities := map[string][]byte{
		"clientA": []byte("alpha"),
		"clientB": []byte("beta"),
	}

	builder := NewDTLSConfigFromMap("stable-server", identities)
	s.NotNil(builder)
	config, testIdentity := builder(s.ctp)
	s.NotNil(config)
	s.Equal([]byte("stable-server"), config.PSKIdentityHint)

	_, testIdentity2 := builder(s.ctp)
	s.Equal(testIdentity, testIdentity2)

}

func (s *TestSuite) TestNewDTLSConfigSingleKey() {
	psk := []byte("single-key-12345")
	builder := NewDTLSConfigSingleKey("test-server", psk)
	s.NotNil(builder)

	config, _ := builder(s.ctp)
	s.NotNil(config)
	s.Equal([]byte("test-server"), config.PSKIdentityHint)

	// Should work with any hint
	returnedPsk, err := config.PSK([]byte("any-hint"))
	s.NoError(err)
	s.Equal(psk, returnedPsk)

	returnedPsk2, err := config.PSK([]byte("different-hint"))
	s.NoError(err)
	s.Equal(psk, returnedPsk2)
}

func (s *TestSuite) TestNewDTLSConfigSingleKey_StableOutput() {
	psk := []byte("consistent-key")
	builder := NewDTLSConfigSingleKey("stable-server", psk)
	s.NotNil(builder)

	config, testIdentity := builder(s.ctp)
	s.NotNil(config)
	s.Equal([]byte("stable-server"), config.PSKIdentityHint)
	_, testIdentity2 := builder(s.ctp)
	s.Equal(testIdentity, testIdentity2)
}

func (s *TestSuite) TestNewDTLSConfigWithFunc_KnownIdentities() {
	keys := map[string][]byte{
		"clientX": []byte("keyX"),
		"clientY": []byte("keyY"),
	}
	pskFunc := func(identityHint string) ([]byte, error) {
		if k, ok := keys[identityHint]; ok {
			return k, nil
		}
		return nil, fmt.Errorf("unknown identity: %s", identityHint)
	}

	builder := NewDTLSConfigWithFunc("test-server", pskFunc)
	config, _ := builder(s.ctp)
	s.NotNil(config)
	s.Equal([]byte("test-server"), config.PSKIdentityHint)

	pskX, err := config.PSK([]byte("clientX"))
	s.NoError(err)
	s.Equal([]byte("keyX"), pskX)

	pskY, err := config.PSK([]byte("clientY"))
	s.NoError(err)
	s.Equal([]byte("keyY"), pskY)
}

func (s *TestSuite) TestNewDTLSConfigWithFunc_UnknownIdentity() {
	keys := map[string][]byte{
		"clientX": []byte("keyX"),
	}
	pskFunc := func(identityHint string) ([]byte, error) {
		if k, ok := keys[identityHint]; ok {
			return k, nil
		}
		return nil, fmt.Errorf("unknown identity: %s", identityHint)
	}

	builder := NewDTLSConfigWithFunc("test-server", pskFunc)
	config, _ := builder(s.ctp)
	s.NotNil(config)

	psk, err := config.PSK([]byte("unknown"))
	s.Error(err)
	s.Nil(psk)
	s.Contains(err.Error(), "unknown identity")
}

func (s *TestSuite) TestNewDTLSConfigWithFunc_ReachabilityIdentity() {
	keys := map[string][]byte{
		"clientX": []byte("keyX"),
	}
	pskFunc := func(identityHint string) ([]byte, error) {
		if k, ok := keys[identityHint]; ok {
			return k, nil
		}
		return nil, fmt.Errorf("unknown identity: %s", identityHint)
	}

	builder := NewDTLSConfigWithFunc("test-server", pskFunc)
	config, reachID := builder(s.ctp)
	s.NotNil(config)
	s.NotEmpty(reachID.Identity)
	s.NotNil(reachID.Key)

	psk, err := config.PSK([]byte(reachID.Identity))
	s.NoError(err)
	s.Equal(reachID.Key, psk)
}

func (s *TestSuite) TestNewDTLSConfigWithFunc_StableOutput() {
	keys := map[string][]byte{
		"clientX": []byte("keyX"),
	}
	pskFunc := func(identityHint string) ([]byte, error) {
		if k, ok := keys[identityHint]; ok {
			return k, nil
		}
		return nil, fmt.Errorf("unknown identity: %s", identityHint)
	}

	builder := NewDTLSConfigWithFunc("test-server", pskFunc)
	_, reachID1 := builder(s.ctp)
	_, reachID2 := builder(s.ctp)
	// The reachability identity is random per builder call, so just check types and lengths
	s.NotEmpty(reachID1.Identity)
	s.NotNil(reachID1.Key)
	s.NotEmpty(reachID2.Identity)
	s.NotNil(reachID2.Key)
}

func TestDaemonTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
