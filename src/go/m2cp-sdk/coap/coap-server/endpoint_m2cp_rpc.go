package coap_server

import (
	"fmt"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"m2cp"
	"m2cp/messages"
	"strings"
)

const coapEndpoint = "/m2cp/rpc"

// EndpointRpc factory for an endpoint that receives a serialized m2cp-command and executes it as a proxy, returning the result
func EndpointRpc(node m2cp.Node) m2cp.CoapEndpoint {
	return m2cp.CoapEndpoint{
		Path:        coapEndpoint,
		ContentType: message.AppOctets,
		Handler: func(r m2cp.CoapRequest) error {
			msgSerialized := r.GetBody()
			cmd, err := messages.CommandFromBinary(msgSerialized)
			_ = cmd
			if err != nil {
				return fmt.Errorf("%d - failed deserializing command message: %s", codes.BadRequest, err)
			}

			topic := strings.SplitN(cmd.GetHeader().GetTopic(), "/", 2)
			toIp, err := messages.NewAddress(r.Ctp(), topic[1])
			if err != nil {
				return fmt.Errorf("%d - failed parsing address: %s", codes.BadRequest, err)
			}
			toLocal, err := messages.NewAddress(r.Ctp(), fmt.Sprintf("%s.%s.local", toIp.GetNodeName(), toIp.GetAppName()))
			if err != nil {
				return fmt.Errorf("%d - failed parsing address: %s", codes.BadRequest, err)
			}

			resChan, err := node.RemoteProcedureCallsWithOptions(
				toLocal,
				cmd.GetCommand(),
				cmd.GetParameters(),
				m2cp.RemoteProcedureCallOptions{
					Trace: cmd.GetTraces(),
				},
			)
			if err != nil {
				return err
			}

			var res m2cp.ResponseMessage

			for proxyResReadonly := range resChan {
				res, err = messages.NewResponseMessage(cmd)
				proxyRes := make([]m2cp.RpcResult, len(proxyResReadonly))
				for i, v := range proxyResReadonly {
					proxyRes[i] = v.(m2cp.RpcResult)
				}
				if len(proxyRes) > 0 {
					res.SetTraces(proxyRes[0].GetTraces())
				} else {
					res.AddTrace("zero results - proxy error")
				}
				res.SetResults(proxyRes)
			}

			res.AddTrace(node.GetAddress())

			resBin, err := messages.ToBinary(res, m2cp.MessageSerializerDefault)
			if err != nil {
				return fmt.Errorf("%d - failed serializing response message: %s", codes.InternalServerError, err)
			}
			r.SetResponseBytes(resBin)
			r.SetResponseCode(codes.Content)
			return nil
		},
	}
}
