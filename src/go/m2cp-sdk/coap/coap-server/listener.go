package coap_server

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/plgd-dev/go-coap/v3/dtls"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	coapNet "github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp"
	coap2 "m2cp/coap"
	coap_client "m2cp/coap/coap-client"
	"time"

	"m2cp"
	"sync"
)

// listener implements m2cp.Listener
// only one of listenersDTLS or listenersUDP should be set at any time
type listener struct {
	ctp               m2cp.ContextPlus
	listenerType      m2cp.ListenerType
	listenerInstances []*listenerInstance
}

type listenerInstance struct {
	ctp          m2cp.ContextPlus
	router       *mux.Router
	dtlsConf     m2cp.DTLSConfig
	dtlsListener *coapNet.DTLSListener
	udpListener  *coapNet.UDPConn
	reachable    bool
}

// NewListener creates a new CoAP server listener based on the provided addresses and options.
// After creation the Serve method should be called to start the server.
// The listener must be closed after usage, using the Close method or cancelling the provided context.
func NewListener(ctpParent m2cp.ContextPlus, addresses []string, endpoints []m2cp.CoapEndpoint, options m2cp.CoapServerOptions) (m2cp.Listener, error) {
	if len(addresses) == 0 {
		return nil, errors.New("no addresses provided")
	}

	listenerInstances := make([]*listenerInstance, 0)

	for _, address := range addresses {
		actualAddress := appendStandardPort(address)

		var err error
		lst := &listenerInstance{}
		if options.DTLS != nil {
			ctp := ctpParent.BranchWithName(fmt.Sprintf("server@%s", actualAddress))
			lst.dtlsListener, err = createDTLSListener(ctp, actualAddress, options.DTLS)
			if err != nil {
				return nil, fmt.Errorf("cannot create DTLS listener at %s: %w", actualAddress, err)
			}
			actualAddress = lst.dtlsListener.Addr().String()
		} else {
			lst.udpListener, err = createUDPListener(actualAddress)
			if err != nil {
				return nil, fmt.Errorf("cannot create UDP listener at %s: %w", actualAddress, err)
			}
			actualAddress = lst.udpListener.LocalAddr().String()
		}

		//override ctp with actual address (in case port 0 was used)
		ctp := ctpParent.BranchWithName(fmt.Sprintf("server@%s", actualAddress))
		lst.ctp = ctp
		lst.router = createRouter(ctp, actualAddress, options.VerbosityLevel, endpoints)
		lst.dtlsConf = options.DTLS

		lst.closeOnCancel()

		listenerInstances = append(listenerInstances, lst)
	}
	lType := m2cp.ListenerTypeUDP
	if options.DTLS != nil {
		lType = m2cp.ListenerTypeDTLS
	}

	return listener{
		ctp:               ctpParent.BranchWithName(fmt.Sprintf("server@%s", addresses[0])),
		listenerType:      lType,
		listenerInstances: listenerInstances,
	}, nil
}

func (l listener) Serve() error {

	wg := sync.WaitGroup{}
	for _, lst := range l.listenerInstances {
		wg.Add(1)

		go lst.run()
		go func() {
			lst.reachabilityTest(l.ctp)
			wg.Done()
		}()
	}

	l.ctp.LogDebug("Waiting for server to respond..")
	wg.Wait()

	initSuccess := true
	for _, lst := range l.listenerInstances {
		if !lst.reachable {
			l.ctp.LogError("initialization failed - server %s didn't respond: %v", lst.getAddress())
			initSuccess = false
		}
	}

	if !initSuccess {
		for _, lst := range l.listenerInstances {
			lst.close()
		}
		return errors.New("failed starting CoAP server(s)")
	}

	return nil

}

func (l *listenerInstance) closeOnCancel() {
	if l.dtlsListener != nil {
		go func() {
			<-l.ctp.Done()
			l.ctp.LogDebug("Stopping DTLS listener..")
			errC := l.dtlsListener.Close()
			if errC != nil {
				l.ctp.LogError("Error closing DTLS listener: %v", errC)
			}
		}()
	} else {
		go func() {
			<-l.ctp.Done()
			l.ctp.LogDebug("Stopping UDP listener..")
			errC := l.udpListener.Close()
			if errC != nil {
				l.ctp.LogError("Error closing UDP listener: %v", errC)
			}
		}()
	}
}

func (l *listenerInstance) run() {
	var err error
	defer l.close()
	if l.dtlsListener != nil {
		s := dtls.NewServer(options.WithMux(l.router), options.WithContext(l.ctp))
		err = s.Serve(l.dtlsListener)
	} else if l.udpListener != nil {
		s := udp.NewServer(options.WithMux(l.router), options.WithContext(l.ctp))
		err = s.Serve(l.udpListener)
	} else {
		err = errors.New("no listener available")
	}
	if err != nil {
		l.ctp.LogError("CoAP server stopped with error: %v", err)
	} else {
		l.ctp.LogDebug("CoAP server stopped")
	}
}

func (l listener) Close() {
	for _, lst := range l.listenerInstances {
		lst.close()
	}

}

func (l *listenerInstance) close() {
	l.ctp.Cancel()
}

func (l *listenerInstance) reachabilityTest(ctpParent m2cp.ContextPlus) {
	var client m2cp.CoapClient
	var err error

	address := l.getAddress()
	if address == "" {
		ctpParent.LogError("cannot perform reachability test: no address available")
		l.reachable = false
		return
	}

	if l.dtlsConf == nil {
		client, err = coap_client.NewClient(l.ctp, address)
	} else {
		_, dtlsOpts := l.dtlsConf(l.ctp)
		client, err = coap_client.NewClientWithOptions(l.ctp, address, m2cp.CoapClientOptions{DTLS: coap2.NewDTLSConfigSingleKey(dtlsOpts.Identity, dtlsOpts.Key)})
	}
	if err != nil {
		ctpParent.LogError("creating reachability client for server at %s: %w", address, err)
		return
	}
	for i := 0; i < 10; i++ {
		res, err := client.Get("/.well-known/core", "", nil)
		if err != nil {
			continue
		}
		if res.GetResponseCode() == codes.Content {
			l.reachable = true
			return
		}
		if ctpParent.IsCancelled() {
			return
		}
		ctpParent.Sleep(1 * time.Second)
	}
	ctpParent.LogError("no response from server at %s after 10 attempts", address)
}

func (l listener) GetAddresses() []string {

	var addresses = make([]string, 0)

	if l.listenerType == m2cp.ListenerTypeDTLS {
		for _, lst := range l.listenerInstances {
			if lst == nil {
				continue
			}
			if addr := lst.getAddress(); addr != "" {
				addresses = append(addresses, addr)
			}
		}
	} else if l.listenerType == m2cp.ListenerTypeUDP {
		for _, lst := range l.listenerInstances {
			if lst == nil {
				continue
			}
			if addr := lst.getAddress(); addr != "" {
				addresses = append(addresses, addr)
			}
		}

	}
	return addresses
}

func (l listener) GetAddress(position uint) string {

	if position >= uint(len(l.listenerInstances)) {
		return ""
	} else {
		return l.listenerInstances[position].getAddress()
	}
}

func (l *listenerInstance) getAddress() string {
	if l.dtlsListener != nil {
		return l.dtlsListener.Addr().String()
	}
	if l.udpListener != nil {
		return l.udpListener.LocalAddr().String()
	}
	return ""
}

func createRouter(ctp m2cp.ContextPlus, address string, verbosity int, endpoints []m2cp.CoapEndpoint) *mux.Router {
	endpointsContainer := endpointsContainerType{endpoints: endpoints}
	endpointsContainer.endpoints = append(endpointsContainer.endpoints, newEndpointWellKnownCore(&endpointsContainer))

	router := mux.NewRouter()
	if verbosity >= 0 {
		router.Use(makeLoggingMiddleware(ctp))
	}

	for _, endpoint := range endpointsContainer.endpoints {
		if err := router.Handle(endpoint.Path, mux.HandlerFunc(makeHandler(ctp, address, verbosity, endpoint))); err != nil {
			panic(fmt.Sprintf("Error setting up handler '%s': %s", endpoint.Path, err))
		}
	}
	router.DefaultHandleFunc(makeHandler(ctp, address, verbosity, endpointCatchAll()))
	return router
}

func createDTLSListener(ctp m2cp.ContextPlus, address string, dtlsConf m2cp.DTLSConfig) (*coapNet.DTLSListener, error) {

	config, _ := dtlsConf(ctp)
	l, err := coapNet.NewDTLSListener("udp", address, config)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func createUDPListener(address string) (*coapNet.UDPConn, error) {
	l, err := coapNet.NewListenUDP("udp", address)
	if err != nil {
		return nil, err
	}
	return l, nil
}
