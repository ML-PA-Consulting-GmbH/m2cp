package snapd

import "os"

func (s *TestSuite) TestGetSnapsInstalled() {
	if os.Getenv("CI") == "true" {
		s.T().Skip("Skipping in CI environment because snapd/virtual device is not available")
		return
	}
	installed, err := GetSnapsInstalled(s.ctp)
	s.NoError(err)
	s.NotEmpty(installed)

}
