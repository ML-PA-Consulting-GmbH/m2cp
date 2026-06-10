package coap_server

import (
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"m2cp"
	"sort"
	"strings"
)

// newEndpointWellKnownCore is a conventional endpoint for exploration, as per https://datatracker.ietf.org/doc/html/rfc7252#section-6.1
func newEndpointWellKnownCore(container *endpointsContainerType) m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path: "/.well-known/core",
		Handler: func(r m2cp.CoapRequest) error {
			endpointDescriptions := make([]string, len(container.endpoints))
			for i, endpoint := range container.endpoints {
				endpointDescriptions[i] += fmt.Sprintf(`<%s>;ct="%d"`, endpoint.Path, endpoint.ContentType)
			}
			sort.Strings(endpointDescriptions)
			r.SetResponseString(strings.Join(endpointDescriptions, ","))
			r.SetResponseCode(codes.Content)
			return nil
		},
		ContentType: message.AppLinkFormat,
	}
}
