package snapd

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"m2cp"
	"m2cp/contextplus"
	"m2cp/tests"
	"testing"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (s *TestSuite) SetupSuite() {
	tests.LockResource("m2cp-messaging")
	fmt.Println(">>> From SetupSuite")
}

func (s *TestSuite) TearDownSuite() {
	tests.FreeResource("m2cp-messaging")
	fmt.Println(">>> From TearDownSuite")
}

func (s *TestSuite) SetupTest() {
	fmt.Println("-- From SetupTest")
	s.ctp = contextplus.NewContextPlus()
}

func (s *TestSuite) TearDownTest() {
	fmt.Println("-- From TearDownTest")
	s.ctp.Cancel()
}

func TestCompleteSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
