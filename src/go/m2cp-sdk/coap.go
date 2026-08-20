package m2cp

import (
	"net"
	"time"

	piondtls "github.com/pion/dtls/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
)

type Identity struct {
	Identity string
	Key      []byte
}
type DTLSConfig func(ctp ContextPlus) (*piondtls.Config, Identity)
type CoapServerOptions struct {
	// VerbosityLevel sets the verbosity level for the server instance.
	// -1: don't log incoming calls
	// 0: log only endpoint names
	// n: log endpoint names and n bytes of body
	VerbosityLevel int
	DTLS           DTLSConfig
	Listeners      []Listener
}

type ListenerType int

const (
	ListenerTypeUDP ListenerType = iota
	ListenerTypeDTLS
)

type Listener interface {
	GetAddresses() []string
	// GetAddress returns the address at the given position
	// if position is out of range, an empty string is returned
	GetAddress(position uint) string
	Serve() error
	Close()
}

type CoapRequest interface {
	Ctp() ContextPlus
	GetPath() []string
	GetClientAddr() net.Addr
	GetServerAddr() net.Addr
	GetMethod() codes.Code
	RequireMethod(method codes.Code) error
	GetBody() []byte
	GetParam(key string) (string, error)
	GetParamInt(key string) (int, error)
	GetParamUint(key string) (uint, error)
	GetRouteParam(key string) (string, error)
	GetNoResponseOption() uint8
	SetResponseCode(c codes.Code)
	AddResponseBytes(b []byte)
	SetResponseBytes(b []byte)
	SetResponseString(s string)
	SetResponseJson(object interface{})
	SetResponseError(c codes.Code, m string)
	GetPeer() (CoapPeer, error)
	GetLegacyPeer() CoapPeer
}

type CoapResponse interface {
	GetBody() []byte
	GetResponseCode() codes.Code
	GetRemoteAddress() string
	GetLocalAddress() string
}

// CoapClient is the interface for a CoAP client, which can send requests to a CoAP server
type CoapClient interface {
	Get(endpoint string, query string, payload []byte) (response CoapResponse, err error)
	Put(endpoint string, query string, payload []byte) (response CoapResponse, err error)
	Post(endpoint string, query string, payload []byte) (response CoapResponse, err error)

	// MulticastGet sends a CoAP GET request to a multicast group
	// and returns a list of responses from all peers that responded before the timeout
	// it blocks until the timeout is reached
	MulticastGet(endpoint string, query string) ([]CoapResponse, error)
	// MulticastPut sends a CoAP PUT request to a multicast group
	// and returns a list of responses from all peers that responded before the timeout
	// it blocks until the timeout is reached
	MulticastPut(endpoint string, query string, payload []byte) ([]CoapResponse, error)
	// MulticastPost sends a CoAP POST request to a multicast group
	// and returns a list of responses from all peers that responded before the timeout
	// it blocks until the timeout is reached
	MulticastPost(endpoint string, query string, payload []byte) ([]CoapResponse, error)

	MulticastGetAsync(endpoint string, query string, callback func(resp CoapResponse)) error
	MulticastPutAsync(endpoint string, query string, payload []byte, callback func(resp CoapResponse)) error
	MulticastPostAsync(endpoint string, query string, payload []byte, callback func(resp CoapResponse)) error

	SetHost(host string)
	SetPort(port string)
	SetNetwork(network string)
	SetRequestTimeout(timeout time.Duration)

	GetHost() string
	GetPort() string
	GetNetwork() string
	GetRequestTimeout() time.Duration
}

// CoapPeer describes a known peer in a CoAP servers cache of known peers
type CoapPeer interface {
	String() string
	HasDetails() bool
	UpdatePeer(newPeer CoapPeer)
	Serialize() ([]byte, error)

	GenerateOSSerial() (string, error)
	GetOSSerial() string

	SetAddress(address string)
	GetAddress() string
	SetEDAddress(edAddress string)
	GetEDAddress() string

	GetHardwareSerial() string
	SetHardwareSerial(serial string)
	GetMCUID() string
	SetMCUID(mcuid string)
	GetHardwareModel() string
	GetHardwareVendor() string

	SetDeviceModel(model string)
	GetDeviceModel() string
	GetAppName() string
	GetAppVersion() string
	SetSequenceNumber(seq int)
	GetSequenceNumber() int

	SetUpdateResult(result int, resultCode string)

	GetLastSeen() time.Time
	SetLastSeenNow()
	GetUptime() int64
	SetUptime(uptime int64)
	AddFirmwareKey(key string)
	SetFirmwareKeys(keys []string)
	GetFirmwareKeys() []string

	GetFirmWareType() int
	GetHardWareRevision() int
	SetLegacy(legacy bool)
	IsLegacy() bool
}

type CoapEndpoint struct {
	Path        string
	Handler     CoapEndpointHandler
	ContentType message.MediaType
	LogLevel    *LogLevel
}

type CoapClientOptions struct {
	TransportProtocol CoapTransportProtocol
	SourceAddr        string
	NoResponse        uint8
	DTLS              DTLSConfig
	MCastInterface    *net.Interface
}

type CoapTransportProtocol string

const (
	TransportProtocolUDP CoapTransportProtocol = "udp"
	// TransportProtocolUDP4 is deprecated, set to TransportProtocolUDP
	TransportProtocolUDP4 CoapTransportProtocol = "udp"
	// TransportProtocolUDP6 is deprecated, set to TransportProtocolUDP
	TransportProtocolUDP6 CoapTransportProtocol = "udp"

	NoResponseIgnore2_xx = 2
	NoResponseIgnore4_xx = 8
	NoResponseIgnore5_xx = 16
	NoResponseIgnoreAll  = NoResponseIgnore2_xx | NoResponseIgnore4_xx | NoResponseIgnore5_xx
)

type CoapEndpointHandler func(r CoapRequest) error
