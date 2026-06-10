package device

// TODO: this should be refactored to use the snapd package

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"m2cp"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func getSnapdSerialFromSocket() (string, error) {
	socketPath := "/run/snapd.socket"

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: time.Second * 10, // or another appropriate timeout
	}

	resp, err := client.Get("http://localhost/v2/assertions/serial")
	if err != nil {
		return "", fmt.Errorf("Error making request: %s\n", err)
	}
	defer func() {
		_ = resp.Body.Close()
		client.CloseIdleConnections()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response body: %s\n", err)
	}

	return snapdSerialFromAssertion(body)
}

func getSnapdSerialFromUrl(ctx m2cp.ContextPlus, port int) (string, error) {
	timeout := 10 * time.Second
	client := &http.Client{
		Timeout: timeout,
	}
	defer client.CloseIdleConnections()

	url := fmt.Sprintf("http://127.0.0.1:%d/v2/assertions/serial", port)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	limitedBody := io.LimitReader(resp.Body, 1<<16) // Limit to 64k
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return "", err
	}

	return snapdSerialFromAssertion(body)
}

func snapdSerialFromAssertion(body []byte) (string, error) {
	var err error

	parts := strings.SplitN(string(body), "\n\n", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected response format")
	}
	assertionBody := parts[0]

	var assertion map[string]interface{}
	err = yaml.Unmarshal([]byte(assertionBody), &assertion)
	if err != nil {
		return "", err
	}

	serial, ok := assertion["serial"]
	if !ok {
		return "", fmt.Errorf("no serial in assertion")
	}

	return fmt.Sprintf("%v", serial), nil
}

type snapCtlModel struct {
	Architecture string `json:"architecture"`
	BrandId      string `json:"brand-id"`
	Classic      string `json:"classic"`
	Model        string `json:"model"`
	Revision     string `json:"revision"`
	Serial       string `json:"serial"`
	//Timestamp    time.Time `json:"timestamp"`
}

func getSnapdSerialFromSnapCtl() (string, error) {
	cmd := exec.Command("snapctl", "model", "--json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running snapctl model: %w", err)
	}

	var model snapCtlModel
	if err = json.Unmarshal(output, &model); err != nil {
		return "", fmt.Errorf("error unmarshalling snapctl model: %w", err)
	}

	if model.Serial == "" {
		return "", fmt.Errorf("no serial in snapctl model output")
	}
	return model.Serial, nil
}
