package uptime

import (
	coap_client "m2cp/coap/coap-client"
	"m2cp/tests"
	"time"
)

// TestUptimeRequest can't really test anything without a running server
// This is more of a usage example than a real test
func (t *TestSuite) TestUptimeRequest() {
	resetRequestedRuntimes()
	RequestAppRuntime(t.ctp, "test", 10)
}

func (t *TestSuite) TestLeakageRunningMockServer() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	t.MockUptimeServer(port)
	time.Sleep(5 * time.Second)
}

func (t *TestSuite) TestLeakageCallingServer() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()
	t.T().Skip("Skipping because it doesn't run and either way there are leaks")

	t.MockUptimeServer(port)
	t.ctp.Sleep(1 * time.Second)

	client, err := coap_client.NewClient(t.ctp, uptimeManagerIp)
	t.NoError(err)
	res, err := client.Put("/request_uptime/1", "app=anonymous_app_test_foo&seconds=42", nil)
	t.NoError(err)
	t.NotNil(res)
}

func (t *TestSuite) TestUptimeRequestMultiple() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	t.MockUptimeServer(port)

	resetRequestedRuntimes()
	RequestAppRuntime(t.ctp, "test1", 30)
	RequestAppRuntime(t.ctp, "test2", 20)
	RequestAppRuntime(t.ctp, "test1", 10)
	t.ctp.Sleep(1 * time.Second)
	t.ctp.LogDebug("identifieres: %v", getRuntimeRequestIdentifiers())
	t.Len(getRuntimeRequestIdentifiers(), 2)

	t.ctp.Sleep(5 * time.Second)
}
