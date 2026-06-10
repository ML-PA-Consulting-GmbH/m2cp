package networks

import (
	"m2cp/device"
	"strings"
)

func (s *TestSuite) TestGetContainerName() {
	name := getContainerName(s.ctp)
	s.True(strings.HasPrefix(name, "development-"))
}

func (s *TestSuite) TestGetMachineName() {
	name, err := device.GetName(s.ctp)
	s.NoError(err)
	s.NotEmpty(name)
}
