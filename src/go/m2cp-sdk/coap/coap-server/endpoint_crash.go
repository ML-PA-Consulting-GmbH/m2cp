package coap_server

import (
	"github.com/plgd-dev/go-coap/v3/message"
	"m2cp"
)

// EndpointCrash is a simple endpoint that crashes
func EndpointCrash() m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path:        "/crash",
		ContentType: message.TextPlain,
		Handler: func(r m2cp.CoapRequest) error {
			a := 0
			b := 1
			x := b / a
			_ = x
			return nil
		},
	}
}
