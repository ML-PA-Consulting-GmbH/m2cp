package coap_client

import (
	"bytes"
	"fmt"
	"io"
	"m2cp"
	"m2cp/coap"
	net2 "net"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/plgd-dev/go-coap/v3/dtls"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/net/blockwise"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp"
	udpclient "github.com/plgd-dev/go-coap/v3/udp/client"
)

type IpAddressVersion int

const (
	IPUnknown = 0
	IPv4      = 4
	IPv6      = 6
)

type client struct {
	ctp             m2cp.ContextPlus
	addrPort        netip.AddrPort
	coapCallTimeout time.Duration
	source          string
	requestOpts     []message.Option
	dtlsConf        m2cp.DTLSConfig
	mCastInterface  *net2.Interface
}

type coapResponse struct {
	code       codes.Code
	body       []byte
	remoteAddr string
	localAddr  string
}

func NewClient(ctpParent m2cp.ContextPlus, host string) (m2cp.CoapClient, error) {

	ctp := ctpParent.BranchWithName("coap-client")

	addrPort, err := coap.ParseHostPort(host)
	if err != nil {
		return nil, err
	}

	return &client{
		ctp:             ctp,
		addrPort:        addrPort,
		coapCallTimeout: time.Second * 10,
	}, nil
}

func NewClientWithOptions(ctpParent m2cp.ContextPlus, host string, opts m2cp.CoapClientOptions) (m2cp.CoapClient, error) {

	ctp := ctpParent.BranchWithName("coap-client")

	addrPort, err := coap.ParseHostPort(host)
	if err != nil {
		return nil, err
	}

	var dtlsConf m2cp.DTLSConfig
	if opts.DTLS != nil {
		dtlsConf = opts.DTLS
	}

	source := ""
	if opts.SourceAddr != "" {
		srcAddrPort, err := coap.ParseHostPort(opts.SourceAddr)
		if err != nil {
			return nil, err
		}
		srcAddr := srcAddrPort.Addr()
		srcPort := srcAddrPort.Port()

		srcHost := srcAddr.String()
		if srcAddr.Is6() {
			srcHost = "[" + srcHost + "]"
		}
		source = fmt.Sprintf("%s:%d", srcHost, srcPort)
	}

	// Override addrPort if custom transport protocol is specified
	if opts.TransportProtocol != "udp" && opts.TransportProtocol != "" && opts.TransportProtocol != "udp4" && opts.TransportProtocol != "udp6" {
		// For non-UDP protocols, we'll need to track this separately
		// This is a special case that doesn't fit the AddrPort model
	}

	requestOpts := []message.Option{}
	if opts.NoResponse != 0 {
		requestOpts = append(requestOpts, message.Option{ID: message.NoResponse, Value: []byte{opts.NoResponse}})
	}

	return &client{
		ctp:             ctp,
		addrPort:        addrPort,
		coapCallTimeout: time.Second * 10,
		source:          source,
		requestOpts:     requestOpts,
		dtlsConf:        dtlsConf,
		mCastInterface:  opts.MCastInterface,
	}, nil
}

func (o *client) SetHost(host string) {
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return // Invalid address, keep existing
	}
	o.addrPort = netip.AddrPortFrom(addr, o.addrPort.Port())
}

func (o *client) SetNetwork(network string) {
	// Network is derived from the address type in addrPort
	// This method is kept for backward compatibility but is a no-op
	// Use NewClientWithOptions with TransportProtocol to override network type
}

func (o *client) SetPort(port string) {
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 0 || portNum > 65535 {
		return // Invalid port, keep existing
	}
	o.addrPort = netip.AddrPortFrom(o.addrPort.Addr(), uint16(portNum))
}

func (o *client) SetRequestTimeout(timeout time.Duration) {
	o.coapCallTimeout = timeout
}

func (o *client) GetHost() string {
	return o.addrPort.Addr().String()
}

func (o *client) GetPort() string {
	return strconv.Itoa(int(o.addrPort.Port()))
}

func (o *client) GetNetwork() string {
	if o.addrPort.Addr().Is6() {
		return "udp6"
	} else if o.addrPort.Addr().Is4() {
		return "udp4"
	}
	return "udp"
}

func (o *client) GetRequestTimeout() time.Duration {
	return o.coapCallTimeout
}

func (o *client) Get(endpoint string, query string, payload []byte) (response m2cp.CoapResponse, err error) {
	return o.coapCall(endpoint, query, "GET", payload)
}

func (o *client) Put(endpoint string, query string, payload []byte) (response m2cp.CoapResponse, err error) {
	return o.coapCall(endpoint, query, "PUT", payload)
}

func (o *client) Post(endpoint string, query string, payload []byte) (response m2cp.CoapResponse, err error) {
	return o.coapCall(endpoint, query, "POST", payload)
}

func (o *client) coapCall(endpoint string, query string, method string, payloadBytes []byte) (response m2cp.CoapResponse, err error) {
	if o.ctp.IsCancelled() {
		return nil, fmt.Errorf("context is cancelled")
	}

	var readSeeker io.ReadSeeker
	if payloadBytes == nil {
		payloadBytes = []byte{}
	}

	// Convert to io.ReadSeeker
	readSeeker = bytes.NewReader(payloadBytes)

	// Now you can use readSeeker wherever an io.ReadSeeker is needed
	// For example, reading from it
	buf := make([]byte, len(payloadBytes))
	_, err = readSeeker.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed converting payload to io.Reader: %s", err)
	}

	var address string
	addr := o.addrPort.Addr()
	port := o.addrPort.Port()
	if addr.Is6() {
		address = fmt.Sprintf("[%s]:%d", addr.String(), port)
	} else if addr.Is4() {
		address = fmt.Sprintf("%s:%d", addr.String(), port)
	} else {
		o.ctp.LogError("Invalid IP address: %s", addr.String())
		return nil, fmt.Errorf("invalid IP address: %s", addr.String())
	}

	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	// Bind to a specific local address if provided by source
	var laddr *net2.UDPAddr
	if o.source != "" {
		laddr, err = net2.ResolveUDPAddr("udp", o.source)
		if err != nil {
			return nil, fmt.Errorf("failed resolving source UDP address %s: %v", o.source, err)
		}
	}
	dialer := &net2.Dialer{
		LocalAddr: laddr,
	}

	errs := func(e error) {}

	con := &udpclient.Conn{}

	if o.dtlsConf == nil {
		con, err = udp.Dial(address, options.WithDialer(dialer), options.WithErrors(errs), options.WithBlockwise(true, blockwise.SZX1024, time.Second*5))
		if err != nil {
			return nil, fmt.Errorf("failed dialing: %v", err)
		}
	} else {
		config, _ := o.dtlsConf(o.ctp)
		con, err = dtls.Dial(address, config, options.WithDialer(dialer), options.WithErrors(errs), options.WithBlockwise(true, blockwise.SZX1024, time.Second*5))
		if err != nil {
			return nil, fmt.Errorf("failed dialing DTLS: %v", err)
		}
	}

	defer func() {
		_ = con.Close()
	}()

	// New context with timeout
	ctpTimeout := o.ctp.BranchWithTimeout(o.coapCallTimeout)
	defer func() {
		_ = con.Close()
		ctpTimeout.Cancel()
	}()

	opts, err := queryToOptions(query)
	if err != nil {
		return nil, err
	}

	for _, opt := range o.requestOpts {
		// Otherwise, add it as a separate option
		opts = append(opts, opt)

	}

	var resp *pool.Message
	switch method {
	case "GET":
		resp, err = con.Get(ctpTimeout, endpoint, opts...)
	case "PUT":
		resp, err = con.Put(ctpTimeout, endpoint, message.AppOctets, readSeeker, opts...)
	case "POST":
		resp, err = con.Post(ctpTimeout, endpoint, message.AppOctets, readSeeker, opts...)
	default:
		fmt.Println("Invalid method:", method)
		os.Exit(1)
	}

	if err != nil {
		return nil, err
	} else if resp == nil {
		err = fmt.Errorf("null response - timeout")
		return nil, err
	}

	response, err = m2cpCoapResponseFromCoapResponse(resp, con)
	if err != nil {
		return nil, fmt.Errorf("failed converting response: %s", err)
	}

	return response, nil
}

func queryToOptions(query string) ([]message.Option, error) {
	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, fmt.Errorf("Error parsing query: %s\n", err.Error())
	}

	var opts []message.Option
	for key, value := range values {
		if len(value) != 1 {
			return nil, fmt.Errorf("Invalid query: %s\n", query)
		}
		opts = append(opts, message.Option{
			ID:    message.URIQuery,
			Value: []byte(fmt.Sprintf("%s=%s", key, value[0])),
		})
	}

	return opts, nil
}

func stripIpBrackets(ipAddress string) string {
	if strings.HasPrefix(ipAddress, "[") && strings.HasSuffix(ipAddress, "]") {
		return ipAddress[1 : len(ipAddress)-1]
	} else {
		return ipAddress
	}
}

func m2cpCoapResponseFromCoapResponse(resp *pool.Message, conn *udpclient.Conn) (m2cp.CoapResponse, error) {

	response := &coapResponse{}

	if resp != nil {
		response.code = resp.Code()
		bodyReader := resp.Body()
		if bodyReader == nil {
			response.body = nil
		} else {
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(bodyReader)
			if err != nil {
				return nil, err
			}
			response.body = buf.Bytes()
		}
	}

	if conn != nil {
		response.remoteAddr = conn.RemoteAddr().String()
		if resp != nil && resp.ControlMessage() != nil && resp.ControlMessage().Dst != nil {
			response.localAddr = resp.ControlMessage().Dst.String()
		}
	}

	return response, nil
}

func (o *coapResponse) GetResponseCode() codes.Code {
	return o.code
}

func (o *coapResponse) GetRemoteAddress() string {

	return o.remoteAddr
}

func (o *coapResponse) GetBody() []byte {
	return o.body
}
func (o *coapResponse) GetLocalAddress() string {
	return o.localAddr
}

func ipAddressVersionOf(ipAddress string) IpAddressVersion {
	// Strip zone identifier (e.g., %eth1) if present
	ipWithoutZone := strings.Split(ipAddress, "%")[0]
	ip := net2.ParseIP(ipWithoutZone)
	if ip == nil {
		return IPUnknown
	} else if ip.To4() != nil {
		return IPv4
	} else {
		return IPv6
	}

}
