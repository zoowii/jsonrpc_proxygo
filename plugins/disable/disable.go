package disable

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

/**
 * DisableMiddleware is a middleware which can disable some jsonrpc methods
 */
type DisableMiddleware struct {
	proxy.MiddlewareAdapter
	next proxy.Middleware
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
	return middleware.NextOnStart()
}

func (middleware *DisableMiddleware) OnConnection(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnection(session)
}

func (middleware *DisableMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *DisableMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}
func (middleware *DisableMiddleware) OnRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	rpcRequest := session.Request
	if middleware.isDisabledRpcMethod(rpcRequest.Method) {
		response := proxy.NewJSONRpcResponse(rpcRequest.Id, nil, proxy.NewJSONRpcResponseError(proxy.RPC_DISABLED_RPC_METHOD, "disabled rpc method", nil))
		session.FillRpcResponse(response)
		return
	}
	return middleware.NextOnJSONRpcRequest(session)
}
func (middleware *DisableMiddleware) OnRpcResponse(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcResponse(session)
}

func (middleware *DisableMiddleware) ProcessRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextProcessJSONRpcRequest(session)
}
