package rpctypes

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TestSuite struct {
	suite.Suite
}

func (s *TestSuite) SetupSuite() {
	fmt.Println(">>> From SetupSuite")
}

func (s *TestSuite) TearDownSuite() {
	fmt.Println(">>> From TearDownSuite")
}

func (s *TestSuite) SetupTest() {
	fmt.Println("-- From SetupTest")
}

func (s *TestSuite) TearDownTest() {
	fmt.Println("-- From TearDownTest")
}

func TestCompleteSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
