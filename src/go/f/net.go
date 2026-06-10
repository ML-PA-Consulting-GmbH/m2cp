package f

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetJson(url string, target any) error {
	resp, err := GetHttp(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s returned status %d: %s", url, resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode JSON from %s: %w", url, err)
	}
	return nil
}

func GetHttp(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %s failed: %w", url, err)
	}
	return resp, nil
}
