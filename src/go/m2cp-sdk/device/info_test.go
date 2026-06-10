package device

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"m2cp"
	"m2cp/contextplus"
	"os"
	"testing"
)

type TestSuite struct {
	suite.Suite
	ctx m2cp.ContextPlus
}

func (s *TestSuite) SetupSuite() {
	s.T().Logf(">>> From SetupSuite")
}

func (s *TestSuite) TearDownSuite() {
	s.T().Logf(">>> From TearDownSuite")
}

func (s *TestSuite) SetupTest() {
	s.T().Logf("-- From SetupTest")
	s.ctx = contextplus.NewContextPlus()
}

func (s *TestSuite) TearDownTest() {
	s.T().Logf("checking for leaks")
	s.ctx.Cancel()
	goleak.VerifyNone(s.T())
	s.T().Logf("-- From TearDownTest: done")
}

func TestSuiteRunner(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestGetSnapdSerialFromSocket() {
	if os.Getenv("CI") == "true" {
		s.T().Skip("Skipping in CI environment because snapd/virtual device is not available")
		return
	}

	serial, err := getSnapdSerialFromSocket()
	s.NoError(err)
	s.NotEmpty(serial)
	id, err := uuid.Parse(serial)
	s.NoError(err)
	s.NotEmpty(id)
	fmt.Println("serial: " + serial)
}

func (s *TestSuite) TestGetModel() {
	if os.Getenv("CI") == "true" {
		s.T().Skip("Skipping in CI environment because snapd/virtual device is not available")
		return
	}
	model, err := GetModel(s.ctx)
	s.NoError(err)
	s.NotEmpty(model)
	fmt.Println("model: " + model)
}

func (s *TestSuite) TestMakeOsVersion() {
	assertion := []byte(`type: model
authority-id: mlpa
revision: 7
series: 16
brand-id: mlpa
model: m2cp-sil0
architecture: arm64
base: core20
classic: false
grade: signed
snaps:
  -
    default-channel: latest/stable
    id: 0707a279d880749d576dcf814a5f0dff
    name: m2cp-sil0-gadget
    type: gadget
  -
    default-channel: latest/stable
    id: 33bb07b4b91eb075ab233391d1cef1e6
    name: m2cp-sil0-kernel
    type: kernel
  -
    default-channel: latest/stable
    id: 6ebf1fd79130484599c23588496cc282
    name: core20
    type: base
  -
    default-channel: latest/stable
    id: 764abfd79ddd204333e6a06a970b6bf5
    name: m2cp-gateway
    type: app
  -
    default-channel: latest/stable
    id: 4d709cc4c21d9177578b595485ec23ae
    name: m2cp-message-hub
    type: app
  -
    default-channel: latest/stable
    id: 084fe0e3b5f2753ca2fe0d657d12d032
    name: m2cp-coap
    type: app
  -
    default-channel: latest/stable
    id: 3f7c468d715dd622bc84a83aca75e0de
    name: m2cp-systest
    type: app
  -
    default-channel: latest/stable
    id: 87d74466dc28923598f8c3a23ce07ceb
    name: m2cp-config-knorr
    type: app
  -
    default-channel: latest/stable
    id: 008eb10d6b2da12f60d6b7253292e511
    name: m2cp-lte-coap
    type: app
  -
    default-channel: latest/stable
    id: 88b7c4a3e02c017aeee3abe6782dfcaa
    name: m2cp-chrony
    type: app
storage-safety: prefer-encrypted
timestamp: 2024-03-05T13:30:42.3000148Z
sign-key-sha3-384: A8DmNGrpSue33SEyy6DnR-86cbLHyKgA8N2OX0aTzif1L3zTC1iwm3FErA-hbFwM

AcLBUgQAAQoABgUCZeceggAAbi0QACB5ooVOm4pyIfnfc9eqwQCwnT+JNuNNhqn9ZFs2Oq6T25tI
8R/9vMyURAl8PrbacjFFL9SRJV8vYlzAN3wQW/iRMZs6tfiePf0v+Y1oQbZ4jH+Cv3srui1rLOrl
U651KWqL+Hyjahq/bH8SEbDvedaNzZaZ1Swx08djPsGej6psXU8XtQzANGKgVqS2KDyf316Ircd2
DeFrwpciRMlHdVJlaVlyDF+R7HfqE2eON26GzkNYaDzbnDtlv/Aoyh1huEM9IDLByI/PgOfaZ70m
YS298q1Mctck9pL2d9AJaWuILX08ulaeQ+8cGePCDR8KPm5st4cq3hZSLRVh7wK9m3XQd0dZPl27
b+z8xD+7HWCCKadlfB4dhcQGIqDmrl2snTsESImLt/Mz1KDwK1yoaWXmmaegKl6fryWas3p8pRf/
LCiS/oHomd0P//xkuwQOiq5h76pc1TcuOfCkKjfRJyWqo3pKjoBo32q0PSp2OIgBn4Dm742QpneJ
lFNmUJHhA2KVXtn8RA9QINzmoqtcgSAlVWfZODrhCAXl2BdOSDmuM5MXfVRRgK8j6qxv1fBT2Bao
E7prcrIS41ki2dBvFKeGjg7Cx2ajBBxzgozjutgG7jVftFVkArbMv0SHkHkOyGuRPgtvm6WGOwZ8
kb5fggO67d7rPodgwCf/rARJ2FWJ
`)
	snaps := map[string]uint{
		"core20":           5,
		"core22":           6,
		"snapd":            17,
		"m2cp-message-hub": 42, // core app, listed in assertion
		"foo-bar":          99, // ignored, not a core app
	}
	version, err := MakeOsVersion(s.ctx, snaps, assertion)
	s.NoError(err)
	s.NotEmpty(version)
	s.Equal("2-205-226-mh42-sd17", version)

	parsed, err := DecodeOsVersion(version)
	s.NoError(err)
	s.NotNil(parsed)

	s.Equal(uint(5), parsed.Apps["core20"])
	s.Equal(uint(6), parsed.Apps["core22"])
	s.Equal(uint(17), parsed.Apps["snapd"])
	s.Equal(uint(42), parsed.Apps["m2cp-message-hub"])
	s.Len(parsed.Apps, 4)
}

/*
TODO: activate this test once virtual devices expose their CoAP ports
func (s *TestSuite) TestGetHwSerial() {
	serial, err := GetHwSerial(s.ctx)
	s.NoError(err)

	// We will only get a non-empty serial on a real device - so for a test we'll have to ignore this check
	// s.NotEmpty(serial)
	fmt.Println("serial: " + serial)

	// NOTE: the coap client has leaks which clean themselves after a while - problem seems to be deep in the library
	s.ctx.Sleep(10 * time.Second)
}
*/
