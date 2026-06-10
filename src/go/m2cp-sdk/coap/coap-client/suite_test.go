package coap_client

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"m2cp"
	"m2cp/coap"
	"m2cp/contextplus"
	"m2cp/tests"
	gonet "net"
	"testing"
	"time"

	piondtls "github.com/pion/dtls/v3"
	"github.com/plgd-dev/go-coap/v3/dtls"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp"
	"github.com/stretchr/testify/suite"
	"github.com/vishvananda/netlink"
)

type TestSuite struct {
	suite.Suite
	ctp m2cp.ContextPlus
}

func (t *TestSuite) SetupSuite() {
	fmt.Println(">>> From SetupSuite")
}

func (t *TestSuite) TearDownSuite() {
	fmt.Println(">>> From TearDownSuite")
}

func (t *TestSuite) SetupTest() {
	fmt.Println("-- From SetupTest")
	t.ctp = contextplus.NewContextPlus()
}

func (t *TestSuite) TearDownTest() {
	fmt.Println("-- From TearDownTest")
	t.ctp.Cancel()
}

func TestDaemonTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (t *TestSuite) TestShutdownDTLSServer() {
	port, freeResource := tests.GetCoapDTLSTestPort()
	defer freeResource()

	handleGet := func(w mux.ResponseWriter, m *mux.Message) {
		t.ctp.LogInfo("Received GET request")

		err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte("GET response")))
		t.NoError(err)
	}

	c, err := NewClientWithOptions(t.ctp, "[::1]:"+port, m2cp.CoapClientOptions{
		DTLS: coap.NewDTLSConfigSingleKey("client", []byte{0xAB, 0xC1, 0x23}),
	})
	t.NoError(err)
	c.SetRequestTimeout(5 * time.Second)

	confBuilder := coap.NewDTLSConfigSingleKey("server", []byte{0xAB, 0xC1, 0x23})
	serverCtp := t.ctp.Branch()
	dtlsConf, _ := confBuilder(serverCtp)

	err = startDTLSTestServer(serverCtp, port, dtlsConf, handleGet)
	t.NoError(err)
	t.ctp.Sleep(1 * time.Second)
	serverCtp.Cancel()
	t.ctp.Sleep(1 * time.Second)

	err = startDTLSTestServer(serverCtp, port, dtlsConf, handleGet)
	t.NoError(err)
	t.ctp.Sleep(1 * time.Second)
	serverCtp.Cancel()
	t.ctp.Sleep(1 * time.Second)

	res, err := c.Get("/", "", nil)
	t.Error(err)
	t.Nil(res)
}

func startSimpleTestServer(ctx m2cp.ContextPlus, port string, handler func(mux.ResponseWriter, *mux.Message)) error {
	loggingMiddleware := func(next mux.Handler) mux.Handler {
		return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
			ctx.LogDebug("ClientAddress %v, %v\n", w.Conn().RemoteAddr(), r.String())
			next.ServeCOAP(w, r)
		})
	}

	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/", mux.HandlerFunc(handler))

	s := udp.NewServer(options.WithMux(r))

	l, err := net.NewListenUDP("udp6", "[::1]:"+port)
	if err != nil {
		return err
	}

	go func() {
		err := s.Serve(l)
		if err != nil {
			ctx.LogError("Server stopped with error: %v", err)
		}
	}()

	// Stop the server and close the listener when the context is cancelled
	go func() {
		<-ctx.Done()
		err := l.Close()
		if err != nil {
			ctx.LogError("Error closing listener: %v", err)
		}

	}()
	return nil

}

func startDTLSTestServer(ctx m2cp.ContextPlus, port string, dtlsConf *piondtls.Config, handler func(mux.ResponseWriter, *mux.Message)) error {

	m := mux.NewRouter()
	err := m.Handle("/", mux.HandlerFunc(handler))
	if err != nil {
		return err
	}

	l, err := net.NewDTLSListener("udp", "[::1]:"+port, dtlsConf)
	if err != nil {
		return err
	}

	s := dtls.NewServer(options.WithMux(m), options.WithContext(ctx))

	go func() {
		err := s.Serve(l)
		if err != nil {
			ctx.LogError("Server stopped with error: %v", err)
		}
	}()

	// Stop the server and close the listener when the context is cancelled
	go func() {
		<-ctx.Done()
		err := l.Close()
		if err != nil {
			ctx.LogError("Error closing listener: %v", err)
		}

	}()
	return nil
}

func runTestServerMulticast(ctx context.Context, exchangeAddr string, port string, iface *gonet.Interface, handler func(mux.ResponseWriter, *mux.Message)) {

	m := mux.NewRouter()
	err := m.Handle("/", mux.HandlerFunc(handler))

	if err != nil {
		log.Println(err)
	}
	m.Handle("/ping", mux.HandlerFunc(
		func(w mux.ResponseWriter, r *mux.Message) {
			log.Println("Received ping request")

			err := w.SetResponse(codes.Content, message.AppOctets, bytes.NewReader([]byte("pong")))

			if err != nil {
				log.Println(err)
			}
		},
	))

	l, err := net.NewListenUDP("udp4", "0.0.0.0:"+port)
	if err != nil {
		log.Println(err)
		return
	}
	//very dirty but enough for test purposes
	if exchangeAddr[0] != '[' {
		exchangeAddr = "[" + exchangeAddr + "]"
	}

	a, err := gonet.ResolveUDPAddr("udp", exchangeAddr+":"+port)
	if err != nil {
		log.Println(err)
		return
	}

	err = l.JoinGroup(iface, a)
	if err != nil {
		log.Printf("cannot JoinGroup(%v, %v): %v", iface, a, err)
	}
	err = l.SetMulticastLoopback(false)
	if err != nil {
		log.Println("Error for eth0: ", err)
		return
	}

	s := udp.NewServer(options.WithMux(m))

	go func() {
		s.Serve(l)
	}()

	// Stop the server and close the listener when the context is cancelled
	go func() {
		<-ctx.Done()
		s.Stop()
		l.Close()

	}()

}

func addVirtualIP(ctp m2cp.ContextPlus, address string) error {
	link, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("failed to find loopback: %w", err)
	}

	addr, err := netlink.ParseAddr(address)
	if err != nil {
		return fmt.Errorf("parse addr %s: %w", address, err)
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("add addr %s: %w", address, err)
	}
	ctp.LogInfo("Added IP: %s", address)
	return nil
}

func removeVirtualIP(ctp m2cp.ContextPlus, address string) error {
	link, err := netlink.LinkByName("lo")
	if err != nil {

		return fmt.Errorf("cleanup: cannot find loopback: %v", err)
	}

	addr, err := netlink.ParseAddr(address)
	if err != nil {
		return fmt.Errorf("cleanup: parse addr %s: %w", address, err)
	}
	err = netlink.AddrDel(link, addr)
	if err != nil {
		return fmt.Errorf("cleanup: failed to remove %s: %v", address, err)
	}
	ctp.LogInfo("Removed IP: %s", address)

	return nil
}

func getMulticastInterface(ctp m2cp.ContextPlus) (*gonet.Interface, error) {

	ifaceName := "eth0"

	iface, err := gonet.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("interface %s not found: %w", ifaceName, err)
	}

	// Check if interface is UP
	if iface.Flags&gonet.FlagUp == 0 {
		return nil, fmt.Errorf("interface %s is not UP", ifaceName)
	}

	// Check if interface supports Multicast
	if iface.Flags&gonet.FlagMulticast == 0 {
		return nil, fmt.Errorf("interface %s does not support multicast", ifaceName)
	}

	//! For now, ignore IPv6 as the CI containers don't have it configured.
	// Check for IPv6 address
	// addrs, err := iface.Addrs()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get addresses for %s: %w", ifaceName, err)
	// }

	// hasIPv6 := false
	// for _, addr := range addrs {
	// 	if ipnet, ok := addr.(*gonet.IPNet); ok {
	// 		if ipnet.IP.To4() == nil {
	// 			hasIPv6 = true
	// 			break
	// 		}
	// 	}
	// }

	// if !hasIPv6 {
	// 	return nil, fmt.Errorf("interface %s does not have IPv6 configured", ifaceName)
	// }

	//ctp.LogInfo("Interface %s is ready for multicast (UP, Multicast, IPv6)", ifaceName)
	return iface, nil
}
