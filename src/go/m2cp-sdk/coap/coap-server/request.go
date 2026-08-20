package coap_server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"m2cp"
	"net"
	"strconv"
	"strings"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
)

// coapRequest is a wrapper around the underlying coap request featuring some convenience methods
// for accessing the request parameters and writing the response.
type coapRequest struct {
	path          []string
	params        map[string]string
	w             mux.ResponseWriter
	req           *mux.Message
	serverAddr    net.Addr
	ctp           m2cp.ContextPlus
	responseCode  codes.Code
	responseBytes []byte
	responseType  message.MediaType
	body          []byte
	noResponse    uint8
}

func newCoapRequest(components []string, params map[string]string, w mux.ResponseWriter, req *mux.Message, serverAddr net.Addr, ctp m2cp.ContextPlus) coapRequest {
	body, _ := req.ReadBody()

	var noResponse uint32

	if req.HasOption(message.NoResponse) {
		noResponse, _ = req.GetOptionUint32(message.NoResponse)
	}

	r := coapRequest{
		path:          components,
		params:        params,
		body:          body,
		responseCode:  codes.Valid,
		responseBytes: make([]byte, 0),
		responseType:  message.AppOctets,
		w:             w,
		req:           req,
		serverAddr:    serverAddr,
		ctp:           ctp,
		noResponse:    uint8(noResponse),
	}
	return r
}

func (r *coapRequest) GetPath() []string {
	return r.path
}

func (r *coapRequest) GetServerAddr() net.Addr {
	return r.serverAddr
}

func (r *coapRequest) Ctp() m2cp.ContextPlus {
	return r.ctp

}

func (r *coapRequest) GetRouteParam(key string) (string, error) {
	if r.req.RouteParams == nil || r.req.RouteParams.Vars == nil {
		return "", fmt.Errorf("route parameter(s) missing")
	}
	if value, ok := r.req.RouteParams.Vars[key]; ok {
		return value, nil
	} else {
		return "", fmt.Errorf("route parameter(s) missing")
	}
}

func (r *coapRequest) GetParam(key string) (string, error) {
	if value, ok := r.params[key]; ok {
		return value, nil
	} else {
		return "", fmt.Errorf("parameter '%s' required", key)
	}
}

func (r *coapRequest) GetParamInt(key string) (value int, err error) {
	var valueStr string
	if valueStr, err = r.GetParam(key); err == nil {
		var parsedValue int64
		parsedValue, err = strconv.ParseInt(valueStr, 10, 32)
		if err != nil {
			return 0, err
		}
		return int(parsedValue), nil
	} else {
		return 0, err
	}
}

func (r *coapRequest) GetParamUint(key string) (value uint, err error) {
	var valueStr string
	if valueStr, err = r.GetParam(key); err == nil {
		var parsedValue uint64
		parsedValue, err = strconv.ParseUint(valueStr, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(parsedValue), nil
	} else {
		return 0, err
	}
}

func (r *coapRequest) GetMethod() codes.Code {
	return r.req.Message.Code()
}

func (r *coapRequest) RequireMethod(method codes.Code) error {
	if r.req.Message.Code() == method {
		return nil
	}
	return fmt.Errorf("%d - method %v required, %v used", int(codes.MethodNotAllowed), method, r.req.Message.Code())
}

func (r *coapRequest) GetNoResponseOption() uint8 {
	return r.noResponse
}

func (r *coapRequest) AddResponseBytes(b []byte) {
	r.responseBytes = append(r.responseBytes, b...)
}

func (r *coapRequest) SetResponseString(s string) {
	r.responseBytes = []byte(s)
	r.responseType = message.TextPlain
}

func (r *coapRequest) SetResponseBytes(b []byte) {
	if b == nil {
		r.responseBytes = make([]byte, 0)
		return
	}
	r.responseBytes = b
}

func (r *coapRequest) SetResponseJson(object interface{}) {
	payload, err := json.Marshal(object)
	if err != nil {
		r.SetResponseCode(codes.InternalServerError)
		return
	}
	r.SetResponseBytes(payload)
}

func (r *coapRequest) SetResponseCode(c codes.Code) {
	r.responseCode = c
}

func (r *coapRequest) SetResponseError(c codes.Code, m string) {
	r.responseCode = c
	r.SetResponseBytes([]byte(m))
	r.responseType = message.TextPlain
}

func (r *coapRequest) GetBody() []byte {
	return r.body
}

func (r *coapRequest) GetClientAddr() net.Addr {
	if r == nil || r.w == nil {
		return nil
	}
	conn := r.w.Conn()
	if conn == nil {
		return nil
	}
	return conn.RemoteAddr()
}

// GetPeer extracts details from a request and returns a CoapPeer object
func (r *coapRequest) GetPeer() (m2cp.CoapPeer, error) {
	return r.getPeer()
}

func (r *coapRequest) getPeer() (*Peer, error) {
	if r == nil || r.w == nil {
		return nil, fmt.Errorf("request is empty")
	}
	conn := r.w.Conn()
	if conn == nil {
		return nil, fmt.Errorf("connection is empty")
	}
	remote := conn.RemoteAddr()
	if remote == nil {
		return nil, fmt.Errorf("remote address is empty")
	}

	host, _, _ := net.SplitHostPort(remote.String())
	edAddress := ""

	if r.req.ControlMessage() != nil {
		if localAddr := r.req.ControlMessage().Dst; localAddr != nil {
			edHost, _, _ := net.SplitHostPort(localAddr.String())
			if !isMulticastAddress(host) {
				edAddress = edHost
			}
		}
	}

	if len(r.body) == 0 {
		return nil, fmt.Errorf("body is empty")
	}

	p, err := deserializePeer(host, edAddress, r.body)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func isMulticastAddress(host string) bool {
	ipWithoutZone := strings.Split(host, "%")[0]
	remoteIP := net.ParseIP(ipWithoutZone)
	if remoteIP == nil {
		return false
	}

	return remoteIP.IsMulticast()
}

func (r *coapRequest) GetLegacyPeer() m2cp.CoapPeer {
	return r.getLegacyPeer()
}

func (r *coapRequest) getLegacyPeer() *Peer {
	var err error
	if r == nil || r.w == nil {
		return nil
	}
	conn := r.w.Conn()
	if conn == nil {
		return nil
	}
	remote := conn.RemoteAddr()
	if remote == nil {
		return nil
	}

	host, _, _ := net.SplitHostPort(remote.String())
	edAddress := ""

	if r.req.ControlMessage() != nil {
		if localAddr := r.req.ControlMessage().Dst; localAddr != nil {
			edHost, _, _ := net.SplitHostPort(localAddr.String())
			if !isMulticastAddress(host) {
				edAddress = edHost
			}
		}
	}

	p := &Peer{
		Address:   host,
		Legacy:    true,
		EDAddress: edAddress,
	}

	if p.FirmWareType, err = r.GetParamInt("fwt"); err != nil {
		p.FirmWareType = -1
	}
	if p.HardWareRevision, err = r.GetParamInt("hwr"); err != nil {
		p.HardWareRevision = -1
	}
	if p.SequenceNumber, err = r.GetParamInt("fwr"); err != nil {
		p.SequenceNumber = -1
	}

	if fwr, err := r.GetParamInt("seq"); err == nil {
		p.SequenceNumber = fwr
	}

	_, _ = p.GenerateOSSerial()

	return p
}

func (r *coapRequest) WriteResponse(w mux.ResponseWriter, req *mux.Message) {

	if err := w.SetResponse(r.responseCode, r.responseType, bytes.NewReader(r.responseBytes)); err != nil {

		if err.Error() != "message not to be sent due to disinterest" {
			r.ctp.LogError("cannot set response (Code: %v, Token: %v): %v", r.responseCode, req.Token(), err)
		}

	}

}
