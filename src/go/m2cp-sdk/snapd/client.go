package snapd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"m2cp"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// client is an adapter for snapd's REST API, which is documented here:
// https://snapcraft.io/docs/snapd-api
type client struct {
	schemeAndHost string // e.g. "http://localhost:20000"
	transport     *http.Transport
	ctp           m2cp.ContextPlus
}

type apiResponse struct {
	Type       string `json:"type"`
	StatusCode int    `json:"status-code"`
	Status     string `json:"status"`
	Result     any    `json:"result,omitempty"`
	Change     string `json:"change,omitempty"`
}

func newClient(ctpParent m2cp.ContextPlus) (*client, error) {
	s := client{
		ctp: ctpParent.BranchWithName("snapd-adapter"),
	}

	if isVirtualDevice() {
		portNumber, err := getVirtualDevicePortNumber()
		if err != nil {
			return nil, fmt.Errorf("could not get port number: %s", err)
		}
		s.schemeAndHost = fmt.Sprintf("http://localhost:%d", portNumber)
		s.transport = http.DefaultTransport.(*http.Transport).Clone()
	} else {
		socketPath := "/run/snapd.socket"
		s.schemeAndHost = "http://unix"
		s.transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		}
	}

	return &s, nil
}

func (o *client) getModelAssertion() (*ModelAssertion, error) {
	res, err := o.apiGet(fmt.Sprintf("/v2/model"))
	if err != nil {
		return nil, fmt.Errorf("could not get model assertion: %s", err)
	}

	return ParseModelAssertion(res)
}

func (o *client) getSnapsInstalled() ([]Snap, error) {
	res, err := o.apiGet(fmt.Sprintf("/v2/snaps"))
	if err != nil {
		return nil, fmt.Errorf("could not get installed snaps: %s", err)
	}

	var result FindResult
	if err = json.Unmarshal(res, &result); err != nil {
		return nil, fmt.Errorf("could not unmarshal snaps: %s", err)
	}

	return result.Result, nil
}

// apiGet sends a GET request to the snapd API and returns the response
func (o *client) apiGet(path string) ([]byte, error) {
	return o.doRequest("GET", path, nil)
}

// apiPost sends a POST request to the snapd API and returns the response
func (o *client) apiPost(path string, parameters map[string]any) ([]byte, error) {
	return o.doRequest("POST", path, parameters)
}

func (o *client) doRequest(method, path string, parameters map[string]any) ([]byte, error) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	requestUrl, err := url.Parse(fmt.Sprintf("%s/%s", o.schemeAndHost, path))
	if err != nil {
		return nil, fmt.Errorf("could not parse URL")
	}

	httpClient := http.Client{
		Transport: o.transport,
	}
	defer httpClient.CloseIdleConnections()

	var response *http.Response
	switch method {
	case "GET":
		response, err = httpClient.Get(requestUrl.String())
	case "POST":
		var body *bytes.Buffer
		if body, err = makePostBody(parameters); err != nil {
			return nil, err
		}
		response, err = httpClient.Post(requestUrl.String(), "application/json", body)
	default:
		err = fmt.Errorf("unsupported HTTP method: %s", method)
	}
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}
	defer func() { _ = response.Body.Close() }()

	return io.ReadAll(response.Body)
}

func (o *client) ParseRestResponse(responseBodyBytes []uint8) (*apiResponse, error) {
	var target apiResponse
	err := json.NewDecoder(bytes.NewReader(responseBodyBytes)).Decode(&target)
	if err != nil {
		return nil, fmt.Errorf("could not decode JSON response: %o", err)
	}
	return &target, nil
}

func makePostBody(parameters map[string]any) (*bytes.Buffer, error) {
	jsonBytes, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(jsonBytes), nil
}

func getVirtualDevicePortNumber() (int, error) {
	var portNumber int
	var err error
	portNumberStr := os.Getenv("M2CP_VIRTUAL_DEVICE")
	if portNumberStr != "" {
		portNumber, err = strconv.Atoi(portNumberStr)
		if err != nil {
			return 0, fmt.Errorf("could not parse port number: %s", err)
		}
	}
	return portNumber, nil
}

// isVirtualDevice returns true, if this application runs in the context of a Virtual Device
func isVirtualDevice() bool {
	result := false
	portNumber, err := getVirtualDevicePortNumber()
	if portNumber != 0 && err == nil {
		result = true
	}
	return result
}

func (r *apiResponse) String() string {
	return fmt.Sprintf("%s: %s (%d), result \"%v\", change \"%s\"",
		r.Type, r.Status, r.StatusCode, r.Result, r.Change)
}
