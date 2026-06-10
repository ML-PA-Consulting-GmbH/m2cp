package coap_server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"m2cp"
	coap_client "m2cp/coap/coap-client"
	"m2cp/tests"
	"math/rand"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// TestMultiBlock provides a binary download that is bigger than on block (download size is ~31kb)
// This allows testing multi-block capabilities with various clients
func (t *TestSuite) TestMultiBlock() {
	//if os.Getenv("CI") == "true" {
	//	s.ctp.LogInfo("Skipping test TestMultiBlock in CI environment")
	//	return
	//}
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	addresses := []string{"[::]:" + port}

	payload, err := base64.StdEncoding.DecodeString(catImageJpegBase64)
	t.NoError(err)
	t.NotNil(payload)

	err = NewServer(t.ctp, addresses, []m2cp.CoapEndpoint{
		endpointStaticDownload("/download/big-64", []byte(catImageJpegBase64)),
		endpointStaticDownload("/download/big", payload),
		endpointStaticDownload("/download/1000", payload[0:1000]),
		endpointStaticDownload("/download/2000-64", []byte(catImageJpegBase64)[0:2000]),
		endpointStaticDownload("/download/3000", payload[0:3000]),
		endpointStaticDownload("/download/tiny", []byte("hello")),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	client, err := coap_client.NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	res, err := client.Get("/download/big", "", nil)
	t.NoError(err)
	t.NotNil(res)
	t.Len(res.GetBody(), len(payload))

	//s.ctp.LogInfo("Now you can manually call the endpoint with any client - then manually stop the server")
	//t.ctp.Sleep(1 * time.Hour)
}

// TestMultiBlockUpload provides a binary download that is bigger than on block (download size is ~31kb)
// This allows testing multi-block capabilities with various clients
func (t *TestSuite) TestMultiBlockUpload() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	addresses := []string{"[::]:" + port}

	payloadJpeg, err := base64.StdEncoding.DecodeString(catImageJpegBase64)
	t.NoError(err)
	t.NotNil(payloadJpeg)

	err = NewServer(t.ctp, addresses, []m2cp.CoapEndpoint{
		endpointUploadEcho("/upload/echo"),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	client, err := coap_client.NewClient(t.ctp, "[::1]:"+port)
	t.NoError(err)
	payloads := [][]byte{
		[]byte("hello"),
		payloadJpeg,
	}
	for _, payload := range payloads {
		fmt.Printf("Uploading %d bytes\n", len(payload))
		res, err := client.Post("/upload/echo", "", payload)
		t.NoError(err)
		t.NotNil(res)
		t.Len(res.GetBody(), len(payload), "Payload length mismatch")
	}

	//s.ctp.LogInfo("Now you can manually call the endpoint with any client - then manually stop the server")
	//s.ctp.Sleep(1 * time.Hour)
}

func (t *TestSuite) Test20SimultaneousDownloads() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()
	t.T().Skip("Skipping Test20SimultaneousDownloads - libcoap coap-client behaving unexpectedly in concurrent version")

	addresses := []string{"[::]:" + port}

	payload, err := base64.StdEncoding.DecodeString(catImageJpegBase64)
	t.NoError(err)
	t.NotNil(payload)

	err = NewServer(t.ctp, addresses, []m2cp.CoapEndpoint{
		endpointStaticDownload("/download/big", payload),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	numRequests := 20
	signalChan := make(chan struct{})
	wg := sync.WaitGroup{}

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(idx int) {

			resp, err := sendCoAPRequestOnSignal(signalChan, "[::1]:"+port, "GET", "/download/big", nil, true)
			if err != nil {
				// Log error with index for debugging
				t.ctp.LogError(fmt.Sprintf("Request %d failed: %v", idx, err))
				t.Fail("Request failed: %v", err)
			} else {
				t.NotNil(resp)
				t.Len(resp, len(payload)+1) // add 1 for the extra newline added by coap-client
			}
			wg.Done()

		}(i)
	}

	t.ctp.Sleep(1 * time.Second)
	close(signalChan)
	wg.Wait()

}

func (t *TestSuite) TestSendCoAPRequestOnSignal() {
	port, freeResource := tests.GetCoapUDPTestPort()
	defer freeResource()

	addresses := []string{"[::]:" + port}

	err := NewServer(t.ctp, addresses, []m2cp.CoapEndpoint{
		endpointUploadEcho("/upload/echo"),
	}, m2cp.CoapServerOptions{})
	t.NoError(err)
	t.ctp.LogInfo("Server started")

	t.ctp.Sleep(1 * time.Second)

	signalChan := make(chan struct{})
	go func() {
		t.ctp.Sleep(2 * time.Second)
		close(signalChan)
	}()

	payload := []byte("Hello, CoAP!")
	response, err := sendCoAPRequestOnSignal(signalChan, "[::1]:"+port, "POST", "/upload/echo", payload, true)
	t.NoError(err)
	t.NotNil(response)
	t.Contains(string(response), "Hello, CoAP!")
}

func sendCoAPRequestOnSignal(signalChan <-chan struct{}, address string, method string, path string, payload []byte, token bool) ([]byte, error) {
	// Check if coap-client is available
	coapClientPath, err := exec.LookPath("coap-client")
	if err != nil {
		return nil, fmt.Errorf("coap-client not found in PATH")
	}

	// Build coap-client command arguments
	var args []string
	args = append(args, "-m", method)
	if token {
		args = append(args, "-T", strconv.FormatUint(uint64(rand.Uint32()), 16))
	}
	if len(payload) > 0 {
		args = append(args, "-e", string(payload))
	}
	// Construct URI
	uri := fmt.Sprintf("coap://%s%s", address, path)
	args = append(args, uri)

	// Prepare command
	cmd := exec.Command(coapClientPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Wait for signal
	<-signalChan

	// Execute command
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("coap-client error: %v, output: %s", err, out.String())
	}
	return out.Bytes(), nil
}
