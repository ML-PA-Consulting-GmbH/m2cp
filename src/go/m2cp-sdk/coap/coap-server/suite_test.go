package coap_server

import (
	"github.com/stretchr/testify/suite"
	"m2cp"
	"m2cp/contextplus"
	"testing"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (t *TestSuite) SetupSuite() {
	t.T().Logf(">>> From SetupSuite")
}

func (t *TestSuite) TearDownSuite() {
	t.T().Logf(">>> From TearDownSuite")
}

func (t *TestSuite) SetupTest() {
	t.T().Logf("-- From SetupTest")
	t.ctp = contextplus.NewContextPlus()
}

func (t *TestSuite) TearDownTest() {
	t.T().Logf("-- From TearDownTest")
	t.ctp.Cancel()
}

func TestDaemonTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
