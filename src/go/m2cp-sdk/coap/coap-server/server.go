package coap_server

import (
	"fmt"
	"m2cp"
	"m2cp/tools"
	"net"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
)

type EndpointHandler func(r m2cp.CoapRequest) error

func NewServer(ctpParent m2cp.ContextPlus, address []string, endpoints []m2cp.CoapEndpoint, options m2cp.CoapServerOptions) (err error) {

	lst, err := NewListener(ctpParent, address, endpoints, options)
	defer func() {
		if err != nil && lst != nil {
			lst.Close()
		}
	}()

	if err != nil {
		return fmt.Errorf("cannot create CoAP server listener: %w", err)
	}

	err = lst.Serve()
	if err != nil {
		return fmt.Errorf("cannot start CoAP server listener: %w", err)
	}
	return nil

}

func appendStandardPort(address string) string {
	if !strings.HasPrefix(address, "[") {
		return fmt.Sprintf("[%s]:5683", address)
	}
	return address
}

type endpointsContainerType struct {
	endpoints []m2cp.CoapEndpoint
}

// makeHandler creates a handler function for a given endpoint - it wraps the actual handler function
// providing the necessary context and error handling. This allows for keeping the actual handler function
// clean and simple.
func makeHandler(ctpParent m2cp.ContextPlus, address string, verbosity int, endpoint m2cp.CoapEndpoint) func(w mux.ResponseWriter, req *mux.Message) {
	ctp := ctpParent.BranchWithName("coap:/" + endpoint.Path)
	if endpoint.LogLevel != nil {
		ctp.LoosenLogLevel(*endpoint.LogLevel)
	}

	return func(w mux.ResponseWriter, req *mux.Message) {
		// capture panic
		defer func() {
			if r := recover(); r != nil {
				ctp.LogError("panic in handler: %s\n%s", r, debug.Stack())
				_ = w.SetResponse(codes.InternalServerError, message.TextPlain, strings.NewReader("Internal Server Error"))
			}
		}()

		var err error

		// parse path
		path, err := req.Path()
		if err != nil {
			_ = w.SetResponse(codes.BadRequest, message.TextPlain, strings.NewReader(err.Error()))
			return
		}
		pathComponents := strings.Split(strings.TrimPrefix(path, "/"), "/")

		// parse params
		queryParams := make(map[string]string)
		queries, _ := req.Queries()
		for _, param := range queries {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) == 2 {
				key, value := parts[0], parts[1]
				queryParams[key] = value
			}
		}

		// build request object
		r := newCoapRequest(pathComponents, queryParams, w, req, getLocalAddr(address, req), ctp)

		// log call
		if verbosity == 0 {
			ctp.LogDebug("from: %s", r.GetClientAddr())
		} else if verbosity > 0 {
			if tools.IsPrintable(r.GetBody()) {
				ctp.LogDebug("from: %s, body: %s", r.GetClientAddr(), limitStrLen(string(r.GetBody()), verbosity))
			} else {
				ctp.LogDebug("from: %s, body: %s", r.GetClientAddr(), limitStrLen(tools.FormatHex(r.GetBody()), verbosity))
			}
		}

		// execute the actual handler
		if err = endpoint.Handler(&r); err != nil {
			// if the error message is in the format "code - message", we use the code as the response code
			// otherwise we generate a 500 Internal Server Error
			errParts := strings.SplitN(err.Error(), " - ", 2)
			var errCode int
			if errCode, err = strconv.Atoi(errParts[0]); len(errParts) == 2 && err == nil {
				ctp.LogDebug("responding with error code: %d\n", errCode)
				r.SetResponseCode(codes.Code(errCode))
				r.SetResponseBytes([]byte(errParts[1]))
			} else {
				responseCode := codes.InternalServerError
				r.SetResponseCode(responseCode)
				r.SetResponseBytes([]byte(err.Error()))
			}
		}
		// write response
		r.responseType = endpoint.ContentType
		r.WriteResponse(w, req)
	}
}

func getLocalAddr(address string, req *mux.Message) net.Addr {
	host, port, _ := net.SplitHostPort(address)

	ipWithoutZone := strings.Split(host, "%")[0]
	hostIP := net.ParseIP(ipWithoutZone)

	if (host == "" || host == "::" || host == "[::]") && req.ControlMessage() != nil {
		dst := req.ControlMessage().Dst
		if dst != nil {
			hostIP = dst
		}
	}

	return &net.UDPAddr{
		IP: hostIP,
		Port: func() int {
			p, _ := strconv.Atoi(port)
			return p
		}(),
	}

}

func makeLoggingMiddleware(ctp m2cp.ContextPlus) func(mux.Handler) mux.Handler {
	return func(next mux.Handler) mux.Handler {
		return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
			ctp.LogDebug("%s -> %s", w.Conn().RemoteAddr(), r.String())
			next.ServeCOAP(w, r)
		})
	}
}

func limitStrLen(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength] + ".."
	}
	return s
}
