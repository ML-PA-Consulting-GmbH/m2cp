package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type implicitFlowStatusResponse struct {
	ProviderName        *string `json:"providerName"`
	ImplicitFlowEnabled bool    `json:"implicitFlowEnabled"`
	EndpointAvailable   bool    `json:"endpointAvailable"`
}

type sessionResponse struct {
	SessionId string `json:"sessionId"`
	ExpiresIn int    `json:"expiresIn"`
}

type tokenResponse struct {
	// JwtToken is used by legacy (non-external-provider) flows.
	JwtToken string `json:"jwtToken"`
	// BearerToken is the field used by Auth0-backed deployments.
	BearerToken string `json:"bearerToken"`
	// AccessToken is used by some other gateway configurations.
	AccessToken string `json:"accessToken"`
}

// token returns whichever of the response's token fields is populated.
// Different identity provider integrations return the token under different
// field names (see comments on tokenResponse).
func (t tokenResponse) token() string {
	if t.BearerToken != "" {
		return t.BearerToken
	}
	if t.JwtToken != "" {
		return t.JwtToken
	}
	return t.AccessToken
}

type httpClient struct {
	Client  *http.Client
	BaseURL string
}

func LoginWithBrowser(ctx context.Context, storeUrl string) (*string, error) {
	client := createHttpClient(storeUrl)

	err := validateStatus(client)
	if err != nil {
		return nil, err
	}

	sessionResp, err := getSession(client)
	if err != nil {
		return nil, err
	}

	go func() {
		time.Sleep(2 * time.Second)

		url := fmt.Sprintf("%sauth/sessions/%s/authorize", client.BaseURL, sessionResp.SessionId)
	_:
		openInBrowser(url)
	}()

	tokenResp, err := getToken(client, *sessionResp)
	if err != nil {
		return nil, err
	}

	token := tokenResp.token()
	if token == "" {
		return nil, fmt.Errorf("login response did not include a token (no jwtToken/bearerToken/accessToken)")
	}

	return &token, nil
}

func createHttpClient(storeUrl string) *httpClient {
	baseUrl := strings.TrimSuffix(storeUrl, "/")
	baseUrl = strings.Replace(strings.ToLower(baseUrl), "/graphql", "/", 1)

	return &httpClient{
		Client:  &http.Client{},
		BaseURL: baseUrl,
	}
}

func validateStatus(httpClient *httpClient) error {
	url := httpClient.BaseURL + "auth/status"
	resp, err := httpClient.Client.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var status implicitFlowStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return err
	}

	if !status.ImplicitFlowEnabled {
		return fmt.Errorf("implicit flow is not enabled")
	}

	if !status.EndpointAvailable {
		return fmt.Errorf("implicit flow authorization endpoint is not available now")
	}

	return nil
}

func getSession(httpClient *httpClient) (*sessionResponse, error) {
	url := httpClient.BaseURL + "auth/sessions"
	resp, err := httpClient.Client.Post(url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var sessionResp sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sessionResp); err != nil {
		return nil, err
	}

	return &sessionResp, nil
}

func getToken(httpClient *httpClient, session sessionResponse) (*tokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(session.ExpiresIn)*time.Second)
	defer cancel()

	url := fmt.Sprintf("%sauth/sessions/%s/token", httpClient.BaseURL, session.SessionId)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func openInBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		if isWSL() {
			cmd = exec.Command("cmd.exe", "/c", "start", url)
		} else {
			cmd = exec.Command("xdg-open", url)
		}
	case "windows":
		// Note: "cmd" is the command interpreter for Windows, "/c" tells cmd to execute the string that follows,
		// and "start" is used to open the file or URL.
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser: %v\n", err)
	}

	return nil
}

func isWSL() bool {
	_, exists := os.LookupEnv("WSL_DISTRO_NAME")
	return exists
}
