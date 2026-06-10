package protobuf

import (
	"fmt"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/tests"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	ctp      m2cp.ContextPlus
	pathData string
	port     string
}

func (t *TestSuite) SetupTest() {

	t.T().Logf("-- From SetupTest")
	tests.LockResource("m2cp-messaging")
	t.ctp = contextplus.NewContextPlus()
	if os.Getenv("CI") == "true" {
		t.T().Skip("Skipping in CI environment because snapd/virtual device is not available")
		return
	}
}

func (t *TestSuite) SetupSuite() {
	t.ctp = contextplus.NewContextPlus()
	t.ctp = t.ctp.SetModule("test-suite")
}

func (t *TestSuite) TearDownTest() {
	t.ctp.Cancel()
	tests.FreeResource("m2cp-messaging")

}

func (t *TestSuite) TearDownSuite() {
	fmt.Println("TearDownSuite")
	t.ctp.Cancel()
}

func TestCompleteSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
