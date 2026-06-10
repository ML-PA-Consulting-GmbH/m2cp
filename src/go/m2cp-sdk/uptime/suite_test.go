package uptime

import (
	"github.com/stretchr/testify/suite"
	"m2cp"
	coap_server "m2cp/coap/coap-server"
	"m2cp/contextplus"
	"testing"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (s *TestSuite) SetupSuite() {
	s.T().Logf(">>> From SetupSuite")
}

func (s *TestSuite) TearDownSuite() {
	s.T().Logf(">>> From TearDownSuite")
}

func (s *TestSuite) SetupTest() {
	s.T().Logf("-- From SetupTest")
	s.ctp = contextplus.NewContextPlus()
}

func (s *TestSuite) TearDownTest() {
	s.ctp.Cancel()
	s.T().Logf("checking for leaks")
	s.T().Logf("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!  WARNING - goleak.VerifyNone is deactivated, because server leaks")
	//	goleak.VerifyNone(s.T())
	s.T().Logf("-- From TearDownTest: done")
}

func TestSuiteRunner(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) MockUptimeServer(port string) {
	err := coap_server.NewServer(t.ctp, []string{"[::1]:" + port}, []m2cp.CoapEndpoint{
		{
			Path: "/request_uptime/1",
			Handler: func(r m2cp.CoapRequest) error {
				app, err := r.GetParam("app")
				t.NoError(err)
				t.NotNil(app)
				t.Contains(app, "anonymous_app_")
				t.Contains(app, "_test")

				seconds, err := r.GetParamUint("seconds")
				t.NoError(err)
				t.NotNil(seconds)

				return nil
			},
		},
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
}
