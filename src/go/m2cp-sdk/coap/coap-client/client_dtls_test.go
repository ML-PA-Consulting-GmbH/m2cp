package coap_client

import (
	"bytes"
	"context"
	"m2cp"
	"m2cp/coap"
	"m2cp/tests"
	"time"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/stretchr/testify/assert"
)

func (t *TestSuite) TestDTLSClientGet() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()

	handleGet := func(w mux.ResponseWriter, m *mux.Message) {
		t.ctp.LogInfo("Received GET request")
		err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("GET response")))
		t.NoError(err)
	}

	confBuilder := coap.NewDTLSConfigSingleKey("server", []byte{0xAB, 0xC1, 0x23})

	dtlsConf, _ := confBuilder(t.ctp)
	err := startDTLSTestServer(t.ctp, port, dtlsConf, handleGet)
	t.Require().NoError(err)
	t.ctp.Sleep(1 * time.Second)

	c, err := NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{
		DTLS: coap.NewDTLSConfigSingleKey("client", []byte{0xAB, 0xC1, 0x23}),
	})
	t.NoError(err)
	c.SetRequestTimeout(2 * time.Second)

	res, err := c.Get("/", "", nil)
	t.Require().NoError(err)
	t.Equal("GET response", string(res.GetBody()))
}

func (t *TestSuite) TestDTLSClientPut() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()

	req := ""

	handlePut := func(w mux.ResponseWriter, m *mux.Message) {
		bodyReader := m.Body()
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(bodyReader)

		t.NoError(err)
		req = string(buf.Bytes())
		t.ctp.LogInfo("Received PUT request:" + req)

		err = w.SetResponse(codes.Changed, message.TextPlain, bytes.NewReader([]byte("PUT accepted")))
		t.NoError(err)
	}

	confBuilder := coap.NewDTLSConfigSingleKey("server", []byte{0xAB, 0xC1, 0x23})
	dtlsConf, _ := confBuilder(t.ctp)

	err := startDTLSTestServer(t.ctp, port, dtlsConf, handlePut)
	t.NoError(err)
	t.ctp.Sleep(1 * time.Second)

	c, err := NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{
		DTLS: coap.NewDTLSConfigSingleKey("client", []byte{0xAB, 0xC1, 0x23}),
	})
	t.NoError(err)
	c.SetRequestTimeout(5 * time.Second)

	res, err := c.Put("/", "", []byte("PUT data"))
	t.NoError(err)
	t.Equal("PUT accepted", string(res.GetBody()))
	t.Require().Equal("PUT data", req)
}

func (t *TestSuite) TestDTLSClientPost() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	req := ""

	handleHello := func(w mux.ResponseWriter, m *mux.Message) {

		bodyReader := m.Body()
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(bodyReader)

		t.NoError(err)
		req = string(buf.Bytes())
		t.ctp.LogInfo("Received request:" + req)

		err = w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("Hello, client")))
		t.NoError(err)
	}

	confBuilder := coap.NewDTLSConfigSingleKey("server", []byte{0xAB, 0xC1, 0x23})
	dtlsConf, _ := confBuilder(t.ctp)

	err := startDTLSTestServer(t.ctp, port, dtlsConf, handleHello)
	t.NoError(err)
	t.ctp.Sleep(1 * time.Second)

	c, err := NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{DTLS: coap.NewDTLSConfigSingleKey("client", []byte{0xAB, 0xC1, 0x23})})
	t.NoError(err)
	c.SetRequestTimeout(5 * time.Second)

	res, err := c.Post("/", "", []byte("Hello, server"))

	t.Require().NoError(err)
	t.Equal("Hello, client", string(res.GetBody()))
	t.Equal("Hello, server", req)

}

// TestDTLSClientMulticastPut does not differ from TestUDPClientMulticastPut except it uses DTLS configuration.
// As DTLS doesn't work with multicast, the client is expected to use udp instead.
func (t *TestSuite) TestDTLSClientMulticastPut() {

	port, freeResource := tests.GetCoapDTLSTestPort()
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

	c, err := NewClientWithOptions(t.ctp, mCastAddr+":"+port, m2cp.CoapClientOptions{TransportProtocol: m2cp.TransportProtocolUDP6, DTLS: coap.NewDTLSConfigSingleKey("client", []byte{0xAB, 0xC1, 0x23}), MCastInterface: iface})
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
