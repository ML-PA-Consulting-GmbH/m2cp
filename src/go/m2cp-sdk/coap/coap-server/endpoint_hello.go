package coap_server

import (
	"github.com/plgd-dev/go-coap/v3/message"
	"m2cp"
)

// EndpointHello is a simple endpoint that returns "hello!" to any request
func EndpointHello() m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path:        "/hello",
		ContentType: message.TextPlain,
		Handler: func(r m2cp.CoapRequest) error {
			r.SetResponseString("hello!")
			return nil
		},
	}
}
