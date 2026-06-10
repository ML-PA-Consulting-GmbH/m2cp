package coap_server

import (
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"m2cp"
	"time"
)

// NewEndpointRunning is a simple endpoint to check if the server is running
func NewEndpointRunning() m2cp.CoapEndpointHandler {
	timeStarted := time.Now()

	return func(r m2cp.CoapRequest) error {
		type response struct {
			Uptime  int64  `json:"uptime"`
			Started string `json:"started"`
		}
		r.SetResponseJson(response{
			Uptime:  int64(time.Since(timeStarted).Seconds()),
			Started: timeStarted.Format(time.RFC3339),
		})
		r.SetResponseCode(codes.Content)
		return nil
	}

}
