package snapd

import "os"

func (s *TestSuite) TestStopStartService() {
	if os.Getenv("CI") == "true" {
		s.T().Skip("Skipping in CI environment because snapd/virtual device is not available")
		return
	}
	s.NoError(StopService(s.ctp, "m2cp-message-hub"))
	s.NoError(StartService(s.ctp, "m2cp-message-hub"))
}
