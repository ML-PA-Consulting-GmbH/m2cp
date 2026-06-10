package networks

import (
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/networks/amqp"
	"m2cp/snapd"
	"m2cp/tests"
	"os"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (s *TestSuite) SetupSuite() {
	amqp.DebugSender = true
	amqp.DebugSubscriber = true
	s.T().Logf(">>> From SetupSuite")
}

func (s *TestSuite) TearDownSuite() {
	s.T().Logf(">>> From TearDownSuite")
}

func (s *TestSuite) SetupTest() {

	s.T().Logf("-- From SetupTest")
	tests.LockResource("m2cp-messaging")
	s.ctp = contextplus.NewContextPlus()
	if os.Getenv("CI") == "true" {
		s.T().Skip("Skipping in CI environment because snapd/virtual device is not available")
		return
	}
	s.NoError(snapd.StartService(s.ctp, "m2cp-message-hub"), "failed to start m2cp-message-hub")
}

func (s *TestSuite) TearDownTest() {
	s.ctp.Cancel()
	time.Sleep(1 * time.Second) // give some time for the context to cancel
	tests.FreeResource("m2cp-messaging")
	s.T().Logf("checking for leaks")
	goleak.VerifyNone(s.T())
	s.T().Logf("-- From TearDownTest: done")
}

func TestSuiteRunner(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) restartMessageHub() {
	s.T().Logf("----------------------[ restartMessageHub ]----------------------")
	s.ctp.Sleep(2 * time.Second)
	s.ctp.LogInfo("stopping m2cp-message-hub..")
	s.NoError(snapd.StopService(s.ctp, "m2cp-message-hub"))
	s.ctp.Sleep(2 * time.Second)
	s.ctp.LogInfo("starting m2cp-message-hub..")
	s.NoError(snapd.StartService(s.ctp, "m2cp-message-hub"))
	s.ctp.Sleep(2 * time.Second)
}
