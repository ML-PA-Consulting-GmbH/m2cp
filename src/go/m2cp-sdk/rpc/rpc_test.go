package rpc

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"m2cp"
	"m2cp/m2cp_new"
	"testing"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (s *TestSuite) SetupSuite() {
	fmt.Println(">>> From SetupSuite")
}

func (s *TestSuite) TearDownSuite() {
	fmt.Println(">>> From TearDownSuite")
}

func (s *TestSuite) SetupTest() {
	fmt.Println("-- From SetupTest")
	s.ctp = m2cp_new.ContextPlus()
}

func (s *TestSuite) TearDownTest() {
	fmt.Println("-- From TearDownTest")
}

func TestCompleteSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
