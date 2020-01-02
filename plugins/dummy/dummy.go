package dummy

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

type DummyMiddleware struct {

}

func (middleware *DummyMiddleware) Name() string {
	return "dummy"
}

func (middleware *DummyMiddleware) OnConnection(session *proxy.ConnectionSession) (bool, error) {
	return true, nil
}

func (middleware *DummyMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (bool, error) {
	return true, nil
}

func (middleware *DummyMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (bool, error) {
	return true, nil
}
func (middleware *DummyMiddleware) OnJSONRpcRequest(session *proxy.JSONRpcRequestSession) (bool, error) {
	return true, nil
}
func (middleware *DummyMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (bool, error) {
	return true, nil
}

func (middleware *DummyMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (bool, error) {
	return true, nil
}