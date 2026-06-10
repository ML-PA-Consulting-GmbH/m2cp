package amqp

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/tests"
	"testing"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (t *TestSuite) SetupSuite() {
	fmt.Println(">>> From SetupSuite")
}

func (t *TestSuite) TearDownSuite() {
	fmt.Println(">>> From TearDownSuite")
}

func (t *TestSuite) SetupTest() {
	fmt.Println("-- From SetupTest")
	tests.LockResource("m2cp-messaging")
	t.ctp = contextplus.NewContextPlus()
}

func (t *TestSuite) TearDownTest() {
	fmt.Println("-- From TearDownTest")
	tests.FreeResource("m2cp-messaging")
	t.ctp.Cancel()
	defer func() {
		fmt.Println("checking for leaks..")
		goleak.VerifyNone(t.T())
	}()
}

func (t *TestSuite) ResetContextAndCheckLeaks() {
	t.ctp.Cancel()

	goleak.VerifyNone(t.T())
	t.ctp = contextplus.NewContextPlus()
}

func TestCompleteSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
