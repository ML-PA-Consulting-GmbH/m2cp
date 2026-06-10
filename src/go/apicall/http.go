package apicall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"m2cpcli/env"
	"m2cpcli/tools"
	"net/http"
	"net/url"
	"strings"
)

func RequestJson[TResult any](r *http.Request) (*TResult, error) {
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		if urlError, ok := err.(*url.Error); ok {
			return nil, fmt.Errorf("could not connect to %s", urlError.URL)
		}
		return nil, err
	}
	return parseApiCallResultJson[TResult](res)
}

func RequestRaw(r *http.Request) ([]byte, error) {
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	return parseApiCallResultRaw(res)
}

func PostRaw(e *env.RuntimeEnvironment, path string, body []byte, headers *map[string]string) ([]byte, error) {
	req := buildPostRequest(e, path, body, headers)
	return RequestRaw(req)
}

func PostJson[TResult any](e *env.RuntimeEnvironment, path string, body []byte, headers *map[string]string) (*TResult, error) {
	req := buildPostRequest(e, path, body, headers)
	return RequestJson[TResult](req)
}

func GetJson[TResult any](e *env.RuntimeEnvironment, path string, params *map[string]string, headers *map[string]string) (*TResult, error) {
	req := buildGetRequest(e, path, params, headers)
	return RequestJson[TResult](req)
	//res, err := http.Get(path)
	//if err != nil {
	//	return nil, err
	//}
	//return parseApiCallResultJson[TResult](res)
}

func GetRaw(e *env.RuntimeEnvironment, path string, params *map[string]string, headers *map[string]string) ([]byte, error) {
	req := buildGetRequest(e, path, params, headers)
	return RequestRaw(req)
	//res, err := http.Get(path)
	//if err != nil {
	//	return nil, err
	//}
	//return tools.StreamToBytes(res.Body), nil
}

func buildGetRequest(e *env.RuntimeEnvironment, path string, params *map[string]string, headers *map[string]string) *http.Request {
	if params != nil {
		for k, v := range *params {
			path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", k), v)
		}
	}
	path = e.EndpointURL(path)
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Content-Type", "text/json")
	if headers != nil {
		for k, v := range *headers {
			req.Header.Add(k, v)
		}
	}
	return req
}

func buildPostRequest(e *env.RuntimeEnvironment, path string, body []byte, headers *map[string]string) *http.Request {
	if !strings.HasPrefix(path, "http") {
		path = e.EndpointURL(path)
	}
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "text/json")
	if headers != nil {
		for k, v := range *headers {
			req.Header.Add(k, v)
		}
	}
	return req
}

func parseApiCallResultJson[TResult any](res *http.Response) (*TResult, error) {
	if res.StatusCode != http.StatusOK {
		return nil, makeRequestErrorMessage(res)
	}
	resBytes := tools.StreamToBytes(res.Body)
	var resParsed TResult
	err := json.Unmarshal(resBytes, &resParsed)
	return &resParsed, err
}

func parseApiCallResultRaw(res *http.Response) ([]byte, error) {
	if res.StatusCode != http.StatusOK {
		return nil, makeRequestErrorMessage(res)
	}
	resBytes := tools.StreamToBytes(res.Body)
	return resBytes, nil
}

func makeRequestErrorMessage(res *http.Response) error {
	taskId := res.Header.Get("X-Task-Id")
	if taskId != "" {
		taskId = fmt.Sprintf(" (task id: %s)", taskId)
	}
	return fmt.Errorf("%s%s", res.Status, taskId)
}
