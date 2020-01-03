package dummy

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

type DummyMiddleware struct {

}

func (middleware *DummyMiddleware) Name() string {
	return "dummy"
}

func (middleware *DummyMiddleware) OnStart() (err error) {
	return
}

func (middleware *DummyMiddleware) OnConnection(session *proxy.ConnectionSession) (next bool, err error) {
	next = true
	return
}

func (middleware *DummyMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (next bool, err error) {
	next = true
	return
}

func (middleware *DummyMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (next bool, err error) {
	next = true
	return
}
func (middleware *DummyMiddleware) OnJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}
func (middleware *DummyMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}

func (middleware *DummyMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}