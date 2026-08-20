package coap_client

import (
	"bytes"
	"context"
	"m2cp"
	"m2cp/coap"
	"m2cp/contextplus"
	"m2cp/tests"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/stretchr/testify/assert"
)

func (t *TestSuite) TestNewClient() {

	_, err := NewClient(contextplus.NewContextPlus(), "[fe80::df7:2023:1:3%eth1]:5685")
	t.Assert().NoError(err)
}

func (t *TestSuite) TestParseHostPort() {
	tests := []struct {
		name     string
		input    string
		wantAddr string
		wantPort uint16
		wantErr  bool
	}{
		{
			name:     "IPv6 with zone and port",
			input:    "[fe80::df7:2023:1:3%eth1]:5685",
			wantAddr: "fe80::df7:2023:1:3%eth1",
			wantPort: 5685,
			wantErr:  false,
		},
		{
			name:     "IPv6 with zone without port",
			input:    "[fe80::df7:2023:1:3%eth1]",
			wantAddr: "fe80::df7:2023:1:3%eth1",
			wantPort: 5683,
			wantErr:  false,
		},
		{
			name:     "IPv6 with zone no brackets",
			input:    "fe80::df7:2023:1:3%eth1",
			wantAddr: "fe80::df7:2023:1:3%eth1",
			wantPort: 5683,
			wantErr:  false,
		},
		{
			name:     "IPv6 without zone no brackets",
			input:    "fe80::df7:2023:1:3",
			wantAddr: "fe80::df7:2023:1:3",
			wantPort: 5683,
			wantErr:  false,
		},
		{
			name:     "IPv6 without zone with brackets",
			input:    "[fe80::df7:2023:1:3]",
			wantAddr: "fe80::df7:2023:1:3",
			wantPort: 5683,
			wantErr:  false,
		},
		{
			name:     "IPv6 without zone with brackets and port",
			input:    "[fe80::df7:2023:1:3]:5684",
			wantAddr: "fe80::df7:2023:1:3",
			wantPort: 5684,
			wantErr:  false,
		},
		{
			name:     "IPv4 without port",
			input:    "169.254.0.1",
			wantAddr: "169.254.0.1",
			wantPort: 5683, // Default port
			wantErr:  false,
		},
		{
			name:     "IPv4 with port",
			input:    "169.254.0.1:5684",
			wantAddr: "169.254.0.1",
			wantPort: 5684, // Default port
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.T().Run(tt.name, func(t *testing.T) {
			addrPort, err := coap.ParseHostPort(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantAddr, addrPort.Addr().String())
			assert.Equal(t, tt.wantPort, addrPort.Port())
		})
	}
}

func (t *TestSuite) TestNewClientWithOptions() {

	_, err := NewClientWithOptions(contextplus.NewContextPlus(), "[fe80::df7:2023:1:3%eth1]:5685", m2cp.CoapClientOptions{TransportProtocol: m2cp.TransportProtocolUDP6})
	t.Assert().NoError(err)

}

func (t *TestSuite) TestClientSupressError() {
	ctp := contextplus.NewContextPlus()
	c, err := NewClient(ctp, "[::1]:12345")
	if err != nil {
		t.T().Fatalf("Failed to create client: %v", err)
	}
	_, _ = c.Put("/doesntexist", "", nil)
	ctp.Sleep(3 * time.Second)
}

// TODO: What is this test supposed to do? We do not implement this at all.
func (t *TestSuite) TestClientPublicServer() {
	t.T().Skip("Skipping because we don't implement URLs with coap ")
	ctp := contextplus.NewContextPlus()
	c, err := NewClient(ctp, "coap.me")
	t.Assert().NoError(err)
	resp, err := c.Get("/test", "", nil)
	t.Assert().NoError(err)
	t.Assert().NotNil(resp)

}

func (t *TestSuite) TestClientPut() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()
	req := ""

	handler := func(w mux.ResponseWriter, m *mux.Message) {

		bodyReader := m.Body()
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(bodyReader)

		t.NoError(err)
		req = string(buf.Bytes())
		t.ctp.LogInfo("Received request:" + req)

		err = w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("Hello, client")))
		t.NoError(err)
	}

	err := startSimpleTestServer(t.ctp, port, handler)
	t.NoError(err)
	time.Sleep(1 * time.Second)

	c, err := NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	c.SetRequestTimeout(5 * time.Second)

	res, err := c.Put("/", "", []byte("Hello, server"))

	t.NoError(err)
	t.Equal("Hello, client", string(res.GetBody()))
	t.Equal("Hello, server", req)

}

func (t *TestSuite) TestClientTimeout() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	handler := func(w mux.ResponseWriter, m *mux.Message) {
		t.ctp.Sleep(7 * time.Second)
		err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("Hello, client")))
		t.NoError(err)
	}
	err := startSimpleTestServer(t.ctp, port, handler)
	t.NoError(err)
	time.Sleep(1 * time.Second)
	c, err := NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	c.SetRequestTimeout(5 * time.Second)

	start := time.Now()
	_, err = c.Get("/", "", []byte("Hello, server"))
	t.Error(err)

	d := time.Since(start).Seconds()
	t.GreaterOrEqual(d, 5.0)
	t.Less(d, 5.1)

}

func (t *TestSuite) TestClientContextTimeout() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	handler := func(w mux.ResponseWriter, m *mux.Message) {
		t.ctp.Sleep(7 * time.Second)
		err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("Hello, client")))
		t.NoError(err)
	}
	err := startSimpleTestServer(t.ctp, port, handler)
	t.NoError(err)
	time.Sleep(1 * time.Second)

	// Create a context with 3 second timeout
	ctpWithTimeout := t.ctp.BranchWithTimeout(3 * time.Second)
	defer ctpWithTimeout.Cancel()

	c, err := NewClient(ctpWithTimeout, "[::1]:"+port)
	t.NoError(err)
	// Set a longer request timeout - the context timeout should kick in first
	c.SetRequestTimeout(10 * time.Second)

	start := time.Now()
	_, err = c.Get("/", "", []byte("Hello, server"))
	t.Error(err)

	// Should timeout around 3 seconds (context timeout), not 10 seconds (request timeout)
	d := time.Since(start).Seconds()
	t.GreaterOrEqual(d, 3.0)
	t.Less(d, 3.2)

}

func (t *TestSuite) TestClientGetFromDifferentSource() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	// Check if user is root
	if syscall.Geteuid() == 0 {
		err := addVirtualIP(t.ctp, "fd00::100/128")
		t.NoError(err)
		defer removeVirtualIP(t.ctp, "fd00::100/128")
	} else if os.Getenv("CI") == "true" {
		t.ctp.LogWarn("Skipping TestClientGetFromDifferentSource test that requires root privileges in CI environment")
		return
	}

	addr := ""
	handler := func(w mux.ResponseWriter, m *mux.Message) {
		t.ctp.LogInfo("Received request")
		conn := w.Conn()
		if conn != nil {
			remote := conn.RemoteAddr()
			if remote != nil {
				addr, _, _ = net.SplitHostPort(remote.String())
			}
		}
		err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("Hello, client")))
		t.NoError(err)
	}

	startSimpleTestServer(t.ctp, port, handler)
	time.Sleep(1 * time.Second)

	// Create a new client with a different source address that is not [::1]
	c, err := NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{SourceAddr: "[fd00::100]:5644"})
	t.NoError(err)
	c.SetRequestTimeout(5 * time.Second)
	res, err := c.Get("/", "", []byte("Hello, server"))
	t.NoError(err)

	if res == nil {
		t.Fail("Response is nil")
		return
	}
	t.Equal("Hello, client", string(res.GetBody()))
	t.Equal("fd00::100", addr)

}

func (t *TestSuite) TestClientGetFromDifferentSourceEphemeralPort() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	// Check if user is root
	if syscall.Geteuid() == 0 {
		addVirtualIP(t.ctp, "fd00::100/128")
		defer removeVirtualIP(t.ctp, "fd00::100/128")
	}

	addr := ""
	handler := func(w mux.ResponseWriter, m *mux.Message) {
		t.ctp.LogInfo("Received request")
		conn := w.Conn()
		if conn != nil {
			remote := conn.RemoteAddr()
			if remote != nil {
				addr, _, _ = net.SplitHostPort(remote.String())
			}
		}
		err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("Hello, client")))
		t.NoError(err)
	}

	startSimpleTestServer(t.ctp, port, handler)
	time.Sleep(1 * time.Second)

	// Create a new client with a different source address that is not [::1]
	c, err := NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{SourceAddr: "[fd00::100]"})
	t.NoError(err)
	c.SetRequestTimeout(5 * time.Second)
	res, err := c.Get("/", "", []byte("Hello, server"))
	t.NoError(err)

	if res == nil {
		t.Fail("Response is nil")
		return
	}
	t.Equal("Hello, client", string(res.GetBody()))
	t.Equal("fd00::100", addr)

}

func (t *TestSuite) TestClientMulticastGet() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	iface, err := getMulticastInterface(t.ctp)
	t.Require().NoError(err)
	t.Require().NotNil(iface)
	mCastAddr := "224.0.1.0"
	runTestServerMulticast(ctx, mCastAddr, port, iface, func(writer mux.ResponseWriter, m *mux.Message) {})
	time.Sleep(1 * time.Second)

	c, err := NewClientWithOptions(t.ctp, mCastAddr+":"+port, m2cp.CoapClientOptions{TransportProtocol: m2cp.TransportProtocolUDP6, MCastInterface: iface})
	t.NoError(err)
	c.SetRequestTimeout(2 * time.Second)
	start := time.Now()
	// Send a multicast request
	responses, err := c.MulticastGet("/ping", "")
	t.NoError(err)

	for _, response := range responses {
		body := response.GetBody()
		t.ctp.LogInfo("Received response to multicast request from %s :\n%s", response.GetRemoteAddress(), string(body))
		assert.Equal(t.T(), "pong", string(body))
	}

	t.True(time.Since(start).Seconds() >= 2 && time.Since(start).Seconds() < 2.2)
	t.Equal(1, len(responses))

	//t.ctp.LogDebug("Response: ", res)

}

func (t *TestSuite) TestClientMulticastPut() {

	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res := ""
	onReceive := func(w mux.ResponseWriter, m *mux.Message) {
		t.ctp.LogInfo("Received multicast request")

		bodyReader := m.Body()
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(bodyReader)

		t.NoError(err)
		res = string(buf.Bytes())
		t.ctp.LogInfo("Received request:" + res)

		err = w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("Hello, client")))
		t.NoError(err)
	}

	iface, err := getMulticastInterface(t.ctp)
	t.Require().NoError(err)
	t.Require().NotNil(iface)

	mCastAddr := "224.0.1.0"
	runTestServerMulticast(ctx, mCastAddr, port, iface, onReceive)
	// Wait for the server to start
	time.Sleep(1 * time.Second)

	c, err := NewClientWithOptions(t.ctp, mCastAddr+":"+port, m2cp.CoapClientOptions{TransportProtocol: m2cp.TransportProtocolUDP6, MCastInterface: iface})
	t.NoError(err)
	c.SetRequestTimeout(2 * time.Second)

	start := time.Now()
	responses, err := c.MulticastPut("/", "", []byte("Hello Mcast"))
	t.NoError(err)

	for _, response := range responses {
		body := response.GetBody()
		t.ctp.LogInfo("Received response to multicast request from %s :\n%s", response.GetRemoteAddress(), string(body))
		assert.Equal(t.T(), "Hello, client", string(body))
	}

	t.Equal("Hello Mcast", res)

	seconds := time.Since(start).Seconds()
	t.ctp.LogInfo("Time taken: %f ", seconds)

	t.GreaterOrEqual(seconds, 2.0)
	t.LessOrEqual(seconds, 2.2)
	t.Equal(1, len(responses))

	//t.ctp.LogDebug("Response: ", res)

}
