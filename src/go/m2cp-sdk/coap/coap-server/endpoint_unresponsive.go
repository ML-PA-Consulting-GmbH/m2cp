package coap_server

import (
	"github.com/plgd-dev/go-coap/v3/message"
	"m2cp"
	"time"
)

// EndpointUnresponsive is a simple endpoint that crashes
func EndpointUnresponsive() m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path:        "/unresponsive",
		ContentType: message.TextPlain,
		Handler: func(r m2cp.CoapRequest) error {
			r.Ctp().Sleep(time.Minute)
			return nil
		},
	}
}
