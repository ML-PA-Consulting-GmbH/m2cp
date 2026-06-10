package coap_server

import (
	"fmt"
	"m2cp"
	"m2cp/coap/coap-client"
	"m2cp/messages"
	"m2cp/networks"
	"m2cp/rpc"
	"m2cp/rpc/rpctypes"
	"m2cp/tests"
	"net"
	"os"
	"time"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
)

func (t *TestSuite) TestListeningApproachEphemeralPort() {

	lst, err := NewListener(t.ctp, []string{"[::]:0"}, []m2cp.CoapEndpoint{
		EndpointHello(),
	}, m2cp.CoapServerOptions{
		VerbosityLevel: -1,
	})

	t.Require().NoError(err)

	_, port, _ := net.SplitHostPort(lst.GetAddresses()[0])
	err = lst.Serve()

	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	//Now we need to send a CoAP request to the server
	coapClient, err := coap_client.NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	res, err := coapClient.Get("/hello", "", nil)
	t.NoError(err)
	t.NotNil(res)
	if res != nil {
		t.Equal("hello!", string(res.GetBody()))
	}
}

func (t *TestSuite) TestSimpleCoapServer() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		EndpointHello(),
	}, m2cp.CoapServerOptions{
		VerbosityLevel: -1,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	//Now we need to send a CoAP request to the server
	coapClient, err := coap_client.NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	res, err := coapClient.Get("/hello", "", nil)
	t.NoError(err)
	t.NotNil(res)
	if res != nil {
		t.Equal("hello!", string(res.GetBody()))
	}
}

func (t *TestSuite) TestCatchAllEndpoint() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	ctp := t.ctp.BranchWithName("test-catch-all-endpoint")
	//ctp.SetLogLevel(m2cp.LogLevelFatal)
	err := NewServer(ctp, []string{"[::1]:" + port}, []m2cp.CoapEndpoint{}, m2cp.CoapServerOptions{})

	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)
	//Now we need to send a CoAP request to the server
	coapClient, err := coap_client.NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	res, err := coapClient.Get("/hello", "", nil)
	t.NoError(err)
	t.NotNil(res)
	if res != nil {
		t.Equal(res.GetResponseCode(), codes.NotFound)
	}

}

func (t *TestSuite) TestCrashEndpoint() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	err := NewServer(t.ctp, []string{"[::1]:" + port}, []m2cp.CoapEndpoint{
		EndpointCrash(),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	//Now we need to send a CoAP request to the server
	coapClient, err := coap_client.NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	res, err := coapClient.Get("/crash", "", nil)
	t.NoError(err)
	t.Equal("Internal Server Error", string(res.GetBody()))
}

func (t *TestSuite) TestUnresponsiveEndpoint() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	err := NewServer(t.ctp, []string{"[::1]:" + port}, []m2cp.CoapEndpoint{
		EndpointUnresponsive(),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	// We expect the server to take way too long to respond - and our client to time out before
	coapClient, err := coap_client.NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	coapClient.SetRequestTimeout(1 * time.Second)
	t.T().Logf("calling unresponsive endpoint..")
	res, err := coapClient.Get("/unresponsive", "", nil)
	t.T().Logf("done with call")
	t.Error(err)
	t.Nil(res)
}

func (t *TestSuite) TestWellKnownCoreEndpoint() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	err := NewServer(t.ctp, []string{"[::1]:" + port}, []m2cp.CoapEndpoint{
		EndpointHello(),
		EndpointCrash(),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	//Now we need to send a CoAP request to the server
	coapClient, err := coap_client.NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	res, err := coapClient.Get("/.well-known/core", "", nil)
	t.NoError(err)
	t.Equal(`</.well-known/core>;ct="40",</crash>;ct="0",</hello>;ct="0"`, string(res.GetBody()))
}

// TestMultiIp tests the servers behaviour when being reachable on multiple IPs.
// A bug reported, that the server answers from a different IP than the one it received the request on in that case.
// This test should ensure that the server answers from the same IP as the request was received on.
// To prepare this test, you need to add an additional IP to your network interface:
// $ sudo ip -6 addr add fd12:3456:789a:1::1/128 dev lo
func (t *TestSuite) TestMultiIp() {
	if os.Getenv("CI") == "true" {
		t.ctp.LogInfo("Skipping test TestMultiIp in CI environment")
		return
	}
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	//addresses := []string{"fd12:3456:789a:1::1", "fd12:3456:789a:1::2", "[::1]:" + port}
	//addresses := []string{"fd12:3456:789a:1::1", "fd12:3456:789a:1::2"}
	addresses := []string{"[::]:" + port}

	err := NewServer(t.ctp, addresses, []m2cp.CoapEndpoint{
		EndpointHello(),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	//Now we need to send a CoAP request to the server
	for _, address := range addresses {
		coapClient, err := coap_client.NewClient(t.ctp, address)
		t.NoError(err)
		res, err := coapClient.Get("/hello", "", nil)
		t.NoError(err, "Error sending CoAP request - please make sure you have added the IP fd12:3456:789a:1::1 to your network interface ($ sudo ip -6 addr add fd12:3456:789a:1::1/128 dev lo)")
		t.NotNil(res)
		if res != nil {
			t.Equal("hello!", string(res.GetBody()))
		}
	}

	//t.ctp.Sleep(1 * time.Hour)
}

func (t *TestSuite) TestNoResponseOption() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	addresses := []string{"[::]:" + port}

	err := NewServer(t.ctp, addresses, []m2cp.CoapEndpoint{
		endpointUploadEcho("/upload/echo"),
		endpointErrorResult("/error/result", codes.InternalServerError),
	}, m2cp.CoapServerOptions{})

	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	client, err := coap_client.NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{NoResponse: m2cp.NoResponseIgnoreAll})
	t.NoError(err)
	client.SetRequestTimeout(2 * time.Second)

	_, err = client.Post("/upload/echo", "", []byte("hello"))
	t.Error(err, "Expected error due to NoResponse option")

	_, err = client.Get("/error/result", "", nil)
	t.Error(err, "Expected error due to NoResponse option")

}

func (t *TestSuite) TestNoResponseOptionIgnoreSuccessAnd5_xx() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	addresses := []string{"[::]:" + port}

	err := NewServer(t.ctp, addresses, []m2cp.CoapEndpoint{
		endpointUploadEcho("/upload/echo"),
		endpointErrorResult("/error/internal-error", codes.InternalServerError),
		endpointErrorResult("/error/request-error", codes.BadRequest),
	}, m2cp.CoapServerOptions{})

	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	client, err := coap_client.NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{NoResponse: m2cp.NoResponseIgnore5_xx | m2cp.NoResponseIgnore2_xx})
	t.NoError(err)
	client.SetRequestTimeout(2 * time.Second)

	_, err = client.Post("/upload/echo", "", []byte("hello"))
	t.Error(err, "Expected error due to NoResponse option")

	_, err = client.Get("/error/internal-error", "", nil)
	t.Error(err, "Expected error due to NoResponse option")

	res, err := client.Get("/error/request-error", "", nil)
	t.NoError(err, "Expected no error due to NoResponse option only ignoring 5.xx and 2.xx responses")
	t.NotNil(res, "Expected a response despite NoResponse option")
	if res != nil {
		t.Equal(codes.BadRequest, res.GetResponseCode(), "Expected BadRequest response code")
	}

}

func (t *TestSuite) TestNoResponseDoNotIgnoreOnEmptyOrZeroOption() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	addresses := []string{"[::]:" + port}

	err := NewServer(t.ctp, addresses, []m2cp.CoapEndpoint{
		endpointUploadEcho("/upload/echo"),
	}, m2cp.CoapServerOptions{})

	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	client, err := coap_client.NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{NoResponse: 0})
	t.NoError(err)
	client.SetRequestTimeout(2 * time.Second)

	res, err := client.Post("/upload/echo", "", []byte("hello"))
	t.NoError(err)
	t.NotNil(res)
	if res != nil {
		t.Equal("hello", string(res.GetBody()))
	}

	client, err = coap_client.NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{})
	t.NoError(err)
	client.SetRequestTimeout(2 * time.Second)

	res, err = client.Post("/upload/echo", "", []byte("hello"))
	t.NoError(err)
	t.NotNil(res)
	if res != nil {
		t.Equal("hello", string(res.GetBody()))
	}

}

// TestRpcViaCoap tests the RPC functionality via CoAP
// - nodeCaller emits an RPC call to a target with a network address
// - SDK will route the call to the network address using CoAP
// - the CoAP server routes the call to the target node using its proxy node
// - the target node executes the RPC command and returns the result to the proxy
// - the proxy returns the result to the caller via CoAP
func (t *TestSuite) TestRpcViaCoap() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	// the callee node runs the RPC handler
	conRpcApp, err := networks.NewNetworkConnection(t.ctp)
	t.NotNil(conRpcApp)
	t.NoError(err)

	nodeRpcEndpoint, err := conRpcApp.NewNode("rpc-endpoint")
	t.NotNil(nodeRpcEndpoint)
	t.NoError(err)
	handler, err := rpc.NewRpcHandler(exampleCommand())
	t.NoError(err)
	nodeRpcEndpoint.SetRpcHandler(handler)

	conRpcApp.AwaitReady()

	// the proxy node runs the CoAP server
	conProxy, err := networks.NewNetworkConnection(t.ctp)
	t.NotNil(conProxy)
	t.NoError(err)

	nodeProxy, err := conProxy.NewNode("rpc-proxy")
	t.NotNil(nodeProxy)
	t.NoError(err)
	nodeProxy.SetRpcTimeout(5 * time.Second)

	conProxy.AwaitReady()

	err = NewServer(t.ctp, []string{"[::1]:" + port}, []m2cp.CoapEndpoint{
		EndpointRpc(nodeProxy),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)

	// the caller node emits the RPC call
	conCaller, err := networks.NewNetworkConnection(t.ctp)
	t.NotNil(conCaller)
	t.NoError(err)

	nodeCaller, err := conCaller.NewNode("rpc-caller")
	t.NotNil(nodeCaller)
	t.NoError(err)
	// TODO: reduce timout to 5 seconds
	nodeCaller.SetRpcTimeout(500 * time.Second)

	conCaller.AwaitReady()

	// do the RPC call

	recipient, err := messages.NewAddress(t.ctp, fmt.Sprintf("%s.%s.[::1]:"+port,
		nodeRpcEndpoint.GetNodeName(), nodeRpcEndpoint.GetAppName()))
	t.NoError(err)
	resChan, err := nodeCaller.RemoteProcedureCall(recipient, "Test", map[string]string{
		"x": "42",
	})
	t.NoError(err)

	for res := range resChan {
		t.True(res.IsSuccessful())
		t.ctp.LogInfo("RPC successful: %s", res.GetMessage())
		t.ctp.LogInfo("Result value: %d", res.GetInt("foo", 0))
		traces := res.GetTraces()
		for i, trace := range traces {
			t.ctp.LogInfo("Trace #%d: %s \t%s", i, trace.GetRelay(), trace.GetTimeString())
		}
	}
}

func exampleCommand() m2cp.RpcCommand {
	cmd, err := rpc.NewRpcCommand(
		"Test",
		"test",
		"test",
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeInt, "x", "int-param", false),
		},
		[]m2cp.RpcParameter{
			rpc.NewRpcParameter(m2cp.TypeInt, "y", "int-res", true),
		},
		func(ctp m2cp.ContextPlus, parameters m2cp.RpcParameters) m2cp.RpcResult {
			res := rpctypes.NewRpcResultSuccess("yeah")
			res.SetInt("y", parameters.GetInt("x", 0)+1)
			return res
		})
	if err != nil {
		panic(err)
	}
	return cmd
}

func endpointUploadEcho(path string) m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path:        path,
		ContentType: message.AppOctets,
		Handler: func(r m2cp.CoapRequest) error {
			r.SetResponseBytes(r.GetBody())
			return nil
		},
	}
}

func endpointStaticDownload(path string, payload []byte) m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path:        path,
		ContentType: message.AppOctets,
		Handler: func(r m2cp.CoapRequest) error {
			r.SetResponseBytes(payload)
			return nil
		},
	}
}
func endpointErrorResult(path string, code codes.Code) m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path:        path,
		ContentType: message.TextPlain,
		Handler: func(r m2cp.CoapRequest) error {
			r.SetResponseCode(code)
			r.SetResponseBytes([]byte(fmt.Sprintf("%d - %s", code, "error response")))
			return nil
		},
	}

}

const catImageJpegBase64 = `/9j/4AAQSkZJRgABAQEAAAAAAAD/4QAuRXhpZgAATU0AKgAAAAgAAkAAAAMAAAABAAAAAEABAAEA
AAABAAAAAAAAAAD/2wBDAAoHBwkHBgoJCAkLCwoMDxkQDw4ODx4WFxIZJCAmJSMgIyIoLTkwKCo2
KyIjMkQyNjs9QEBAJjBGS0U+Sjk/QD3/2wBDAQsLCw8NDx0QEB09KSMpPT09PT09PT09PT09PT09
PT09PT09PT09PT09PT09PT09PT09PT09PT09PT09PT09PT3/wAARCAHaAdoDASIAAhEBAxEB/8QA
HwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQR
BRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdI
SUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2
t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEB
AQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMi
MoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpj
ZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbH
yMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwD0Ys2TzTg59aYfvU8C
tCRwJ9alQt60xBTwKAJAW9acHb1popRSAfvPrS7m9aZThQA7J9aQlvWlFIRQBG+71qB91WiKjdKa
E0UJUPrVKZG5rUkTrVOVKtMloyZQwPvUBdvWrtwmM1Qk4NMmwm80u88UwmjPSmBJvPrS7j61Hmlz
TAfuPrQXNNFBNK4AXNQu59ac5NQuTSCw13PNV5HNPc9ahkNA0RSMfWq7yGnyPVZyPX60MaQFz1oQ
mozQh5qbjLsZOOalBOarxkcVPGc0wLEZPFTo5GKgHTing0wLAc0HPX0pgNPpAJk+tNJPrSmmmmAF
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`
