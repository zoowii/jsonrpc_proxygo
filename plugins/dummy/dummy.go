package dummy


import (
	"net/http"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

type DummyMiddleware struct {

}

func (middleware *DummyMiddleware) OnConnection(w http.ResponseWriter, r *http.Request) (bool, error) {
	return true, nil
}

func (middleware *DummyMiddleware) OnConnectionClosed(w http.ResponseWriter, r *http.Request) (bool, error) {
	return true, nil
}

func (middleware *DummyMiddleware) OnWebSocketFrame(w http.ResponseWriter, r *http.Request,
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