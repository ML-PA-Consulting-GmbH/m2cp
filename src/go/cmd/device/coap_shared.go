package device

import (
	"context"
	"encoding/base64"
	"fmt"
	coap_server "m2cp/coap/coap-server"
	"m2cpcli/backend"
	"m2cpcli/structs"
)

// CoapResultType holds the response returned by a CoAP request.
type CoapResultType struct {
	Response   []byte `json:"Response"`
	StatusCode string `json:"StatusCode"`
}

// CoapRequest performs a CoAP request to an RTD via its associated ED.
// It is exported so it can be reused by other commands (e.g. device reboot)
func CoapRequest(ctx context.Context, dev *structs.Device, method string, path string, query string, body []byte) (result *CoapResultType, err error) {

	if dev.IsED() {
		return nil, fmt.Errorf("can't send CoAP to Edge Device (not implemented)")
	}
	if !dev.IsRTD() {
		return nil, fmt.Errorf("unknown device type")
	}

	if dev.LastEdgeDevice == nil {
		return nil, fmt.Errorf("no ED is known to be connected to this RTD")
	}

	fmt.Printf("identified most recently connected Edge Device: %s\n", dev.LastEdgeDevice.DeviceSerial)
	fmt.Printf("fetching peers list from ED to resolve CoAP address of RTD\n")

	peers, err := FetchPeersMap(ctx, dev.LastEdgeDevice.DeviceSerial, true)
	if err != nil {
		return nil, fmt.Errorf("failed fetching peers from ED: %s", err)
	}

	var peer coap_server.Peer

	for _, peerWrapper := range peers {
		if peerWrapper.Peer.OSSerial == dev.DeviceSerial {
			peer = peerWrapper.Peer
		}
	}
	if peer.OSSerial == "" {
		return nil, fmt.Errorf("didn't find RTD in peers of the ED - can't route CoAP. Peers: \n%v", peers)
	}

	node := fmt.Sprintf("rpc.m2cp-coap.%s", dev.LastEdgeDevice.DeviceSerial)
	rpcCommand := "executeCoapRequest"
	fmt.Printf("calling '%s' on '%s'\n", rpcCommand, node)

	params := map[string]string{
		"address": peer.Address,
		"method":  method,
		"path":    path,
		"query":   query,
		"body":    string(body),
	}

	res, err := backend.DeviceRpc(ctx, node, rpcCommand, params)
	if err != nil {
		return nil, err
	}
	if res.Error != 0 {
		return nil, fmt.Errorf("CoAP via RPC failed: %s", res.Message)
	}

	result = new(CoapResultType)

	var ok bool
	if responseBase64, ok := res.Result["response"]; !ok {
		return nil, fmt.Errorf("result contains no response")
	} else if result.Response, err = base64.StdEncoding.DecodeString(responseBase64); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err)
	}
	if result.StatusCode, ok = res.Result["statusCode"]; !ok {
		return nil, fmt.Errorf("result contains no status code")
	}

	return result, nil
}
