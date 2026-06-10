package apicall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"m2cpcli/env"
	"m2cpcli/tools"
	"net/http"
)

func GetNonce(e *env.RuntimeEnvironment) ([]byte, error) {
	type resultType struct {
		Nonce string `json:"nonce"`
	}
	api := "/api/v1/snaps/auth/nonces"
	res, err := http.Post(e.EndpointURL(api), "application/json", bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", res.StatusCode)
	}
	var result resultType
	bodyBytes := tools.StreamToBytes(res.Body)
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}
	if len(result.Nonce) == 0 {
		return nil, fmt.Errorf("server returned empty nonce")
	}
	return []byte(result.Nonce), nil
}

type postAuthDeveloperVerifyRequest struct {
	Devices []string `json:"devices,omitempty"`
	Action  string   `json:"action,omitempty"`
}

//func VerifyMacaroon(e *env.Environment, state *env.PersistedState) error {
//	postBody, err := json.Marshal(postAuthDeveloperVerifyRequest{
//		Action:  "m2cpdev",
//		Devices: []string{},
//	})
//	if err != nil {
//		return fmt.Errorf("failed to marshal request: %s", err)
//	}
//	api := "/mlpa/auth/developer-verify"
//	request, err := http.NewRequest(http.MethodPost, e.EndpointURL(api), bytes.NewBuffer(postBody))
//	if err != nil {
//		return fmt.Errorf("failed at authentication request: %s", err)
//	}
//	request.Header.Add("Macaroon", state.Macaroon)
//	request.Header.Add("Content-Type", "application/json")
//	res, err := http.DefaultClient.Do(request)
//	if err != nil {
//		if urlError, ok := err.(*url.Error); ok {
//			if opError, ok := urlError.Err.(*net.OpError); ok {
//				if dnsError, ok := opError.Err.(*net.DNSError); ok {
//					return fmt.Errorf("could not connect to store (DNSError: %s)", dnsError.Err)
//				}
//				return fmt.Errorf("could not connect to store (OpError: %s)", opError.Err)
//			}
//			return fmt.Errorf("could not connect to store (urlError: %s)", urlError.URL)
//		}
//		return fmt.Errorf("unknown error (%s)", err)
//	}
//
//	resBytes := tools.StreamToBytes(res.Body)
//	var resParsed env.PersistedState
//	err = json.Unmarshal(resBytes, &resParsed)
//	if err != nil {
//		return fmt.Errorf("failed to unmarshal response: %s", err)
//	}
//
//	if resParsed.IsAuthenticated() {
//		return nil
//	}
//
//	// see, if we get a better error message
//	type resultType struct {
//		Message string `json:"message"`
//	}
//	var resultMessage resultType
//	err = json.Unmarshal(resBytes, &resultMessage)
//	if err != nil {
//		return err
//	}
//	if resultMessage.Message != "" {
//		return fmt.Errorf("not logged in (message %s)", resultMessage.Message)
//	}
//	return fmt.Errorf("not logged in")
//}
