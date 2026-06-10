package m2cp_new

import (
	"context"
	"m2cp"
	"m2cp/coap"
	coap_client "m2cp/coap/coap-client"
	coap_server "m2cp/coap/coap-server"
	"m2cp/contextplus"
	"m2cp/messages"
	"m2cp/networks"
	"m2cp/protobuf"
)

func Address(ctp m2cp.ContextPlus, address string) (m2cp.Address, error) {
	return messages.NewAddressWithSubtopic(ctp, address, "")
}

func AddressWithSubtopic(ctp m2cp.ContextPlus, address, subtopic string) (m2cp.Address, error) {
	return messages.NewAddressWithSubtopic(ctp, address, subtopic)
}

func ContextPlus() m2cp.ContextPlus {
	return contextplus.NewContextPlus()
}

func ContextPlusFromContext(ctx context.Context) m2cp.ContextPlus {
	return contextplus.FromContext(ctx)
}

func NetworkConnection(ctp m2cp.ContextPlus) (m2cp.NetworkConnection, error) {
	return networks.NewNetworkConnection(ctp)
}

func NetworkConnectionWithOptions(ctp m2cp.ContextPlus, options m2cp.NetworkConnectionOptions) (m2cp.NetworkConnection, error) {
	return networks.NewNetworkConnectionWithOptions(ctp, options)
}

// CoapServer launch a new CoAP server
func CoapServer(ctp m2cp.ContextPlus, addresses []string, endpointsList []m2cp.CoapEndpoint) error {
	return coap_server.NewServer(ctp, addresses, endpointsList, m2cp.CoapServerOptions{})
}

// CoapServerDTLS creates and starts a new CoAP server supporting DTLS as secure transport.
func CoapServerDTLS(ctp m2cp.ContextPlus, addresses []string, endpoints []m2cp.CoapEndpoint, dtlsConfig m2cp.DTLSConfig, options m2cp.CoapServerOptions) error {
	if dtlsConfig != nil {
		options.DTLS = dtlsConfig
	}
	return coap_server.NewServer(ctp, addresses, endpoints, options)
}

func CoapServerWithOptions(ctp m2cp.ContextPlus, addresses []string, endpointsList []m2cp.CoapEndpoint, options m2cp.CoapServerOptions) error {
	return coap_server.NewServer(ctp, addresses, endpointsList, options)
}

func CoapClient(ctp m2cp.ContextPlus, host string) (m2cp.CoapClient, error) {
	return coap_client.NewClient(ctp, host)
}

func CoapClientWithOptions(ctp m2cp.ContextPlus, host string, options m2cp.CoapClientOptions) (m2cp.CoapClient, error) {
	return coap_client.NewClientWithOptions(ctp, host, options)
}

func ProtobufParserService(ctpParent m2cp.ContextPlus, con m2cp.NetworkConnection, parsers []m2cp.ProtobufOidParser) (m2cp.ProtobufParserService, error) {
	return protobuf.NewProtobufParserService(ctpParent, con, parsers)
}

func DTLSConfigSingleKey(preSharedKey []byte, serverIdentityHint string) m2cp.DTLSConfig {
	return coap.NewDTLSConfigSingleKey(serverIdentityHint, preSharedKey)
}

func DTLSConfigFromFile(serverIdentityHint string, pskFilePath string) (m2cp.DTLSConfig, error) {
	return coap.NewDTLSConfigFromFile(serverIdentityHint, pskFilePath)
}

func DTLSConfigWithFunc(serverIdentityHint string, pskFunc func(identityHint string) ([]byte, error)) m2cp.DTLSConfig {
	return coap.NewDTLSConfigWithFunc(serverIdentityHint, pskFunc)
}

func DTLSConfigFromMap(serverIdentityHint string, identities map[string][]byte) m2cp.DTLSConfig {
	return coap.NewDTLSConfigFromMap(serverIdentityHint, identities)
}
