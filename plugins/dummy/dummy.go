package dummy

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

type DummyMiddleware struct {
	proxy.MiddlewareAdapter
}

func (middleware *DummyMiddleware) Name() string {
	return "dummy"
}

func (middleware *DummyMiddleware) OnStart() (err error) {
	return middleware.NextOnStart()
}

func (middleware *DummyMiddleware) OnConnection(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnection(session)
}

func (middleware *DummyMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *DummyMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}
func (middleware *DummyMiddleware) OnRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcRequest(session)
}
func (middleware *DummyMiddleware) OnRpcResponse(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcResponse(session)
}

func (middleware *DummyMiddleware) ProcessRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextProcessJSONRpcRequest(session)
}