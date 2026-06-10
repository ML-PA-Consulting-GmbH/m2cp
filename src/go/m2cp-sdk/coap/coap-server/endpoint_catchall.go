package coap_server

import (
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"m2cp"
	"strings"
)

// endpointCatchAll for handling all requests, that can't be handled by the other endpoints
func endpointCatchAll() m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path:        "*",
		ContentType: message.TextPlain,
		Handler: func(r m2cp.CoapRequest) error {
			r.Ctp().LogWarn(fmt.Sprintf("endpoint 'catchall' path: %s", r.GetPath()), m2cp.SignalLevelInfo)
			if strings.Contains(r.GetPath()[len(r.GetPath())-1], "?") {
				return fmt.Errorf("%d - badly formatted request contains query parameters in path", codes.BadRequest)
			}
			return fmt.Errorf("%d - not found", codes.NotFound)
		},
	}
}
