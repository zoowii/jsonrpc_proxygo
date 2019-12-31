package ws_upstream

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"net/http"
)

type WsUpstreamMiddleware struct {
	TargetEndpoint string
}

// TODO

func NewWsUpstreamMiddleware(targetEndpoint string) *WsUpstreamMiddleware {
	return &WsUpstreamMiddleware{
		TargetEndpoint:targetEndpoint,
	}
}

func (middleware *WsUpstreamMiddleware) OnConnection(w http.ResponseWriter, r *http.Request) (bool, error) {
	// TODO: start ws auto-reconnect connection to target endpoint
	return true, nil
}

func (middleware *WsUpstreamMiddleware) OnConnectionClosed(w http.ResponseWriter, r *http.Request) (bool, error) {
	return true, nil
}

func (middleware *WsUpstreamMiddleware) OnWebSocketFrame(w http.ResponseWriter, r *http.Request,
	messageType int, message []byte) (bool, error) {
	return true, nil
}
func (middleware *WsUpstreamMiddleware) OnJSONRpcRequest(session *proxy.JSONRpcRequestSession) (bool, error) {
	return true, nil
}
func (middleware *WsUpstreamMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (bool, error) {
	return true, nil
}

func (middleware *WsUpstreamMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (bool, error) {
	return true, nil
}