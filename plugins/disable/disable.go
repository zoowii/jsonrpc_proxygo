package disable

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

/**
 * DisableMiddleware is a middleware which can disable some jsonrpc methods
 */
type DisableMiddleware struct {
	rpcMethodsBlacklist map[string]interface{}
}

func NewDisableMiddleware() *DisableMiddleware {
	return &DisableMiddleware{
		rpcMethodsBlacklist: make(map[string]interface{}),
	}
}

func (middleware *DisableMiddleware) AddRpcMethodToBlacklist(methodName string) *DisableMiddleware {
	middleware.rpcMethodsBlacklist[methodName] = true
	return middleware
}

func (middleware *DisableMiddleware) Name() string {
	return "disable"
}

func (middleware *DisableMiddleware) isDisabledRpcMethod(methodName string) bool {
	_, ok := middleware.rpcMethodsBlacklist[methodName]
	return ok
}

func (middleware *DisableMiddleware) OnStart() (err error) {
	return
}

func (middleware *DisableMiddleware) OnConnection(session *proxy.ConnectionSession) (next bool, err error) {
	next = true
	return
}

func (middleware *DisableMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (next bool, err error) {
	next = true
	return
}

func (middleware *DisableMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (next bool, err error) {
	next = true
	return
}
func (middleware *DisableMiddleware) OnJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	rpcRequest := session.Request
	if middleware.isDisabledRpcMethod(rpcRequest.Method) {
		response := proxy.NewJSONRpcResponse(rpcRequest.Id, nil, proxy.NewJSONRpcResponseError(proxy.RPC_DISABLED_RPC_METHOD, "disabled rpc method", nil))
		session.FillRpcResponse(response)
		next = false
	}
	return
}
func (middleware *DisableMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}

func (middleware *DisableMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}
