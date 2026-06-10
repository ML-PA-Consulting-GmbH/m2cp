package networks

import (
	"time"
)

func (s *TestSuite) TestShutdownNetworkConnectionWithStop() {
	con, err := NewNetworkConnection(s.ctp)
	s.NoError(err)

	s.ctp.Sleep(1000 * time.Millisecond)
	con.Close()

	// dangling goroutines will be detected by testsuite
}

func (s *TestSuite) TestShutdownNetworkConnectionWithContext() {
	con, err := NewNetworkConnection(s.ctp)
	s.NoError(err)
	_ = con

	s.ctp.Sleep(1000 * time.Millisecond)
	s.ctp.Cancel()

	// dangling goroutines will be detected by testsuite
}

func (s *TestSuite) TestPreventMultipleNetworkConnections() {
	con, err := NewNetworkConnection(s.ctp)
	s.NoError(err)
	_ = con

	go func() {
		s.ctp.Sleep(3000 * time.Millisecond)
		s.ctp.LogDebug("closing first network connection")
		con.Close()
	}()

	con2, err := NewNetworkConnection(s.ctp)
	s.NoError(err)
	_ = con2

	s.ctp.Sleep(1000 * time.Millisecond)
	s.ctp.Cancel()

	// dangling goroutines will be detected by testsuite
}

// TestPreventMultipleNetworkConnectionsWithContextCancel tests if creating a second network connection is halted
// until the first is closed by cancelling its context.
func (s *TestSuite) TestPreventMultipleNetworkConnectionsWithContextCancel() {

	ctx1 := s.ctp.Branch()
	con, err := NewNetworkConnection(ctx1)
	s.NoError(err)
	s.NotNil(con)

	go func() {
		s.ctp.Sleep(3000 * time.Millisecond)
		s.ctp.LogDebug("closing first network connection")
		ctx1.Cancel()
	}()

	con2, err := NewNetworkConnection(s.ctp)
	s.NoError(err)
	_ = con2

	s.ctp.Sleep(1000 * time.Millisecond)
	s.ctp.Cancel()

	// dangling goroutines will be detected by testsuite
}
