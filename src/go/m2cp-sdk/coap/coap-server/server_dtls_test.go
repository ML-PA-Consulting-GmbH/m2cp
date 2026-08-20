package coap_server

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"m2cp"
	"m2cp/coap"
	coap_client "m2cp/coap/coap-client"
	"m2cp/messages"
	"m2cp/networks"
	"m2cp/rpc"
	"m2cp/tests"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/plgd-dev/go-coap/v3/message/codes"
)

type LibcoapDTLSRequest struct {
	Addr     string //required
	Path     string //required
	Method   string //required
	Token    bool   //optional
	Identity string //required
	Key      []byte //required
	Payload  []byte //optional
}

func (t *TestSuite) TestDTLSSimpleCoapServer() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})
	t.ctp.SetLogLevel(m2cp.LogLevelDebug)

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		EndpointHello(),
	}, m2cp.CoapServerOptions{
		VerbosityLevel: 50,
		DTLS:           conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/hello",
		Method:   "get",
		Token:    true,
		Identity: "m2cp-client",
		Key:      []byte{0xAB, 0xC1, 0x23},
	})

	t.NoError(err)
	t.Equal("hello!", string(res))
}

func (t *TestSuite) TestDTLSServerIdentityFile() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf, err := coap.NewDTLSConfigFromFile("m2cp-coap", "../testdata/test-identity.txt")
	t.NoError(err)

	err = NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		EndpointHello(),
	}, m2cp.CoapServerOptions{
		VerbosityLevel: -1,
		DTLS:           conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	idBytes, err := os.ReadFile("../testdata/test-identity.txt")
	t.Require().NoError(err)

	keys, err := coap.ParsePSKString(string(idBytes))
	t.Require().NoError(err)

	//use the second key from the file to verify that identity/key pairs are correctly parsed and not only the first one
	key, ok := keys["test-client-2"]
	t.Require().True(ok)

	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/hello",
		Method:   "get",
		Token:    true,
		Identity: "test-client-2",
		Key:      key,
	})

	t.NoError(err)
	t.Equal([]byte("hello!"), res)
}

func (t *TestSuite) TestDTLSServerWithFunc() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigWithFunc("m2cp-coap", func(identity string) ([]byte, error) {
		if identity == "m2cp-client" {
			return []byte{0xAB, 0xC1, 0x23}, nil
		}
		return nil, fmt.Errorf("unknown identity: %s", identity)
	})

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		EndpointHello(),
	}, m2cp.CoapServerOptions{
		VerbosityLevel: -1,
		DTLS:           conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)
	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/hello",
		Method:   "get",
		Token:    true,
		Identity: "m2cp-client",
		Key:      []byte{0xAB, 0xC1, 0x23},
	})

	t.NoError(err)
	t.Equal([]byte("hello!"), res)

}

func (t *TestSuite) TestDTLSCatchAllEndpoint() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/hello",
		Method:   "get",
		Token:    true,
		Identity: "m2cp-client",
		Key:      []byte{0xAB, 0xC1, 0x23},
	})

	t.Equal(string(res), "4.04 not found")
}

func (t *TestSuite) TestDTLSCrashEndpoint() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		EndpointCrash(),
	}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/crash",
		Method:   "get",
		Token:    true,
		Identity: "m2cp-client",
		Key:      []byte{0xAB, 0xC1, 0x23},
	})

	t.NoError(err)
	t.Equal("5.00 Internal Server Error", string(res))
}

func (t *TestSuite) TestDTLSWellKnownCoreEndpoint() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		EndpointHello(),
		EndpointCrash(),
	}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/.well-known/core",
		Method:   "get",
		Token:    true,
		Identity: "m2cp-client",
		Key:      []byte{0xAB, 0xC1, 0x23},
	})

	t.NoError(err)
	t.Equal([]byte(`</.well-known/core>;ct="40",</crash>;ct="0",</hello>;ct="0"`), res)
}

func (t *TestSuite) TestDTLSPostEndpoint() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		endpointUploadEcho("/upload/echo"),
	}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	payload := []byte("hello from POST")
	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/upload/echo",
		Method:   "post",
		Token:    true,
		Identity: "m2cp-client",
		Key:      []byte{0xAB, 0xC1, 0x23},
		Payload:  payload,
	})

	t.NoError(err)
	t.Equal(payload, res)
}

func (t *TestSuite) TestDTLSPutEndpoint() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		endpointUploadEcho("/upload/echo"),
	}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	payload := []byte("hello from PUT")
	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/upload/echo",
		Method:   "put",
		Token:    true,
		Identity: "m2cp-client",
		Key:      []byte{0xAB, 0xC1, 0x23},
		Payload:  payload,
	})

	t.NoError(err)
	t.Equal(payload, res)
}

func (t *TestSuite) TestDTLSErrorResult() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

	err := NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		endpointErrorResult("/error/result", codes.InternalServerError),
	}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	res, err := libcoapDTLSClientRequest(t.ctp, LibcoapDTLSRequest{
		Addr:     "[::1]:" + port,
		Path:     "/error/result",
		Method:   "get",
		Token:    true,
		Identity: "m2cp-client",
		Key:      []byte{0xAB, 0xC1, 0x23},
	})

	t.NoError(err)
	t.Contains(string(res), "5.00")
	t.Contains(string(res), "error response")
}

func (t *TestSuite) TestDTLSLargePayloadPost() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

	catImageBytes, err := base64.StdEncoding.DecodeString(catImageJpegBase64)
	t.Require().NoError(err)

	err = NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		endpointUploadEcho("/upload/echo"),
	}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	coapClient, err := coap_client.NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{
		DTLS: coap.NewDTLSConfigSingleKey("m2cp-client", []byte{0xAB, 0xC1, 0x23}),
	})
	t.NoError(err)
	res, err := coapClient.Post("/upload/echo", "", catImageBytes)

	t.NoError(err)
	t.NotNil(res)
	if res != nil {
		t.Equal(catImageBytes, res.GetBody())
	}
}

func (t *TestSuite) TestDTLSLargePayloadDownload() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

	catImageBytes, err := base64.StdEncoding.DecodeString(catImageJpegBase64)
	t.Require().NoError(err)

	err = NewServer(t.ctp, []string{"[::]:" + port}, []m2cp.CoapEndpoint{
		endpointStaticDownload("/download/cat.jpg", catImageBytes),
	}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)
	t.ctp.LogInfo("Server started")
	t.ctp.Sleep(1 * time.Second)

	coapClient, err := coap_client.NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{
		DTLS: coap.NewDTLSConfigSingleKey("m2cp-client", []byte{0xAB, 0xC1, 0x23}),
	})
	t.NoError(err)
	res, err := coapClient.Get("/download/cat.jpg", "", nil)

	t.NoError(err)
	t.NotNil(res)
	if res != nil {
		t.Equal(catImageBytes, res.GetBody())
	}
}

// TestDTLSRpcViaCoap tests the RPC functionality via CoAP
// - nodeCaller emits an RPC call to a target with a network address
// - SDK will route the call to the network address using CoAP
// - the CoAP server routes the call to the target node using its proxy node
// - the target node executes the RPC command and returns the result to the proxy
// - the proxy returns the result to the caller via CoAP
func (t *TestSuite) TestDTLSRpcViaCoap() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()
	conf := coap.NewDTLSConfigSingleKey("m2cp-coap", []byte{0xAB, 0xC1, 0x23})

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
	}, m2cp.CoapServerOptions{
		DTLS: conf,
	})
	t.NoError(err)

	// the caller node emits the RPC call
	conCaller, err := networks.NewNetworkConnection(t.ctp)
	t.NotNil(conCaller)
	t.NoError(err)

	nodeCaller, err := conCaller.NewNode("rpc-caller")
	t.NotNil(nodeCaller)
	t.NoError(err)
	nodeCaller.SetRpcTimeout(500 * time.Second)

	conCaller.AwaitReady()

	// do the RPC call
	recipient, err := messages.NewAddress(t.ctp, fmt.Sprintf("%s.%s.[::1]:"+port,
		nodeRpcEndpoint.GetNodeName(), nodeRpcEndpoint.GetAppName()))
	t.NoError(err)
	resChan, err := nodeCaller.RemoteProcedureCallsWithOptions(recipient, "Test", []map[string]string{
		{"x": "42"},
	}, m2cp.RemoteProcedureCallOptions{DTLS: coap.NewDTLSConfigSingleKey("m2cp-client", []byte{0xAB, 0xC1, 0x23})})
	t.NoError(err)

	for results := range resChan {
		for _, res := range results {
			t.True(res.IsSuccessful())
			t.ctp.LogInfo("RPC successful: %s", res.GetMessage())
			t.ctp.LogInfo("Result value: %d", res.GetInt("foo", 0))
			traces := res.GetTraces()
			for i, trace := range traces {
				t.ctp.LogInfo("Trace #%d: %s \t%s", i, trace.GetRelay(), trace.GetTimeString())
			}
		}
	}
}

func libcoapDTLSClientRequest(ctp m2cp.ContextPlus, request LibcoapDTLSRequest) ([]byte, error) {
	coapClientPath, err := exec.LookPath("coap-client")
	if err != nil {
		return nil, errors.New("coap-client not found in PATH")
	}

	//verify that coap-client supports tinyDTLS
	cmd := exec.Command(coapClientPath, "-h")
	var helpOut bytes.Buffer
	cmd.Stdout = &helpOut
	cmd.Stderr = &helpOut

	_ = cmd.Run()

	if !bytes.Contains(helpOut.Bytes(), []byte("TLS Library: TinyDTLS")) {
		return nil, errors.New("coap-client does not support TinyDTLS - DTLS requests cannot be performed")
	}

	// Build coap-client command arguments
	var args []string
	args = append(args, "-m", request.Method)
	if request.Token {
		args = append(args, "-T", strconv.FormatUint(uint64(rand.Uint32()), 16))
	}

	args = append(args, "-B", "10") // set timeout to 10 seconds
	//args = append(args, "-v", "7")  // set verbosity to max for debugging

	if request.Payload != nil {
		if len(request.Payload) > 1024 {
			args = append(args, "-b", "1024")
		}
		args = append(args, "-e", string(request.Payload))
	}
	// Construct URI
	uri := fmt.Sprintf("coaps://%s%s", request.Addr, request.Path)
	args = append(args, uri)

	// DTLS options
	args = append(args, "-u", request.Identity)
	keyHex := fmt.Sprintf("0x%x", request.Key)
	args = append(args, "-k", keyHex)

	// log command once
	ctp.LogDebug("Executing coap-client command: %s %v", coapClientPath, args)

	// Prepare command
	cmd = exec.Command(coapClientPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Execute command
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("coap-client error: %w, output: %s", err, out.String())
	}

	// remove trailing newlines from output
	result := out.Bytes()
	result = bytes.TrimSpace(result)

	return result, nil
}
