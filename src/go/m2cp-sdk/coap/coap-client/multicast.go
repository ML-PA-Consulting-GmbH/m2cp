package coap_client

import (
	"bytes"
	"errors"
	"fmt"
	"m2cp"
	"strings"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/udp"
	udpclient "github.com/plgd-dev/go-coap/v3/udp/client"
)

func (o *client) MulticastGet(endpoint string, query string) ([]m2cp.CoapResponse, error) {
	return o.multicastCall(endpoint, query, "GET", nil)
}

func (o *client) MulticastPut(endpoint string, query string, payload []byte) ([]m2cp.CoapResponse, error) {
	return o.multicastCall(endpoint, query, "PUT", payload)
}

func (o *client) MulticastPost(endpoint string, query string, payload []byte) ([]m2cp.CoapResponse, error) {
	return o.multicastCall(endpoint, query, "POST", payload)
}

func (o *client) MulticastGetAsync(endpoint string, query string, callback func(resp m2cp.CoapResponse)) error {
	return o.multicastCallAsync(endpoint, query, "GET", nil, callback)
}

func (o *client) MulticastPutAsync(endpoint string, query string, payload []byte, callback func(resp m2cp.CoapResponse)) error {
	return o.multicastCallAsync(endpoint, query, "PUT", payload, callback)
}

func (o *client) MulticastPostAsync(endpoint string, query string, payload []byte, callback func(resp m2cp.CoapResponse)) error {
	return o.multicastCallAsync(endpoint, query, "POST", payload, callback)
}

func (o *client) multicastCall(endpoint string, query string, method string, payloadBytes []byte) ([]m2cp.CoapResponse, error) {
	// Ensure endpoint format
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	opts, err := queryToOptions(query)
	if err != nil {
		return nil, err
	}

	var address string
	addr := o.addrPort.Addr()
	port := o.addrPort.Port()
	if addr.Is6() {
		address = fmt.Sprintf("[%s]:%d", addr.String(), port)
	} else if addr.Is4() {
		address = fmt.Sprintf("%s:%d", addr.String(), port)
	} else {
		o.ctp.LogFatal("Invalid IP address: %s", addr.String())
		return nil, fmt.Errorf("invalid IP address: %s", addr.String())
	}

	responses := []m2cp.CoapResponse{}
	callback := func(resp m2cp.CoapResponse) {

		if resp != nil {
			responses = append(responses, resp)
		}
	}

	err = o.multicast(endpoint, opts, address, method, payloadBytes, callback)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

func (o *client) multicastCallAsync(endpoint string, query string, method string, payloadBytes []byte, callback func(resp m2cp.CoapResponse)) error {

	// Ensure endpoint format
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	opts, err := queryToOptions(query)
	if err != nil {
		return err
	}

	var address string
	addr := o.addrPort.Addr()
	port := o.addrPort.Port()
	if addr.Is6() {
		address = fmt.Sprintf("[%s]:%d", addr.String(), port)
	} else if addr.Is4() {
		address = fmt.Sprintf("%s:%d", addr.String(), port)
	} else {
		o.ctp.LogFatal("Invalid IP address: %s", addr.String())
		return fmt.Errorf("invalid IP address: %s", addr.String())
	}

	go o.multicast(endpoint, opts, address, method, payloadBytes, callback)

	return nil
}

func (o *client) multicast(endpoint string, opts []message.Option, address string, method string, payloadBytes []byte, callback func(resp m2cp.CoapResponse)) error {

	if o.ctp.IsCancelled() {
		return nil
	}

	ctp := o.ctp.BranchWithTimeout(o.coapCallTimeout)
	// Create a message pool
	messagePool := pool.New(1024, 1600)
	req := messagePool.AcquireMessage(ctp) // Use a separate context
	defer messagePool.ReleaseMessage(req)

	// Generate a token
	token, err := message.GetToken()
	if err != nil {
		o.ctp.LogError("failed to get token: %w", err)
		return err
	}

	// Setup request based on method
	switch method {
	case "GET":
		err = req.SetupGet(endpoint, token, opts...)
	case "POST":
		err = req.SetupPost(endpoint, token, message.AppOctets, bytes.NewReader(payloadBytes), opts...)
	case "PUT":
		err = req.SetupPut(endpoint, token, message.AppOctets, bytes.NewReader(payloadBytes), opts...)
	default:
		o.ctp.LogError("invalid method: %s", method)
		return err
	}

	if err != nil {
		o.ctp.LogError("failed to create \"%s\" request: %w", method, err)
		return err
	}

	req.SetMessageID(message.GetMID())
	req.SetType(message.NonConfirmable)

	// Create UDP listener
	l, err := net.NewListenUDP(o.GetNetwork(), ":0")
	if err != nil {
		o.ctp.LogError("failed to create UDP listener for incoming responses: %w", err)
		return err
	}

	// Create CoAP server
	s := udp.NewServer()

	// Start the CoAP server in a goroutine

	go func() {
		o.ctp.LogDebug("Starting CoAP server for receiving multicast responses")
		if err := s.Serve(l); err != nil {
			o.ctp.LogError("Failed to start CoAP server for receiving responses: " + err.Error())
			return
		}
		o.ctp.LogDebug("CoAP multicast response receiver stopped")
	}()

	var mCastOpts []net.MulticastOption
	if o.mCastInterface != nil {
		mCastOpts = append(mCastOpts, net.WithMulticastInterface(*o.mCastInterface))
	}

	// Send the multicast request and handle responses
	// Function returns after req.ctp is cancelled. Either by timeout or if parent context o.ctp is cancelled.
	err = s.DiscoveryRequest(req, address, func(cc *udpclient.Conn, resp *pool.Message) {
		o.ctp.LogDebug("Received response to multicast request")

		response, err := m2cpCoapResponseFromCoapResponse(resp, cc)

		if err != nil {
			o.ctp.LogError("Failed to parse Coap Response into internal structure: %s", err.Error())
			return
		}

		callback(response)

	}, mCastOpts...)
	if err != nil {
		return errors.New("Failed to send multicast request: " + err.Error())
	}
	s.Stop()
	if err := l.Close(); err != nil {

		return errors.New("Failed to close UDP listener: " + err.Error())
	}

	return nil
}
