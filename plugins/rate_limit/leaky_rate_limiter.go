package rate_limit

// TODO: 漏桶算法的限流中间件

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

type LeakyRateLimiterMiddleware struct {
	proxy.MiddlewareAdapter
}

func NewLeakyRateLimiterMiddleware() *LeakyRateLimiterMiddleware {
	return &LeakyRateLimiterMiddleware{}
}

func (middleware *LeakyRateLimiterMiddleware) Name() string {
	return "leaky_rate_limiter"
}

func (middleware *LeakyRateLimiterMiddleware) OnStart() (err error) {
	return middleware.NextOnStart()
}

func (middleware *LeakyRateLimiterMiddleware) OnConnection(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnection(session)
}

func (middleware *LeakyRateLimiterMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *LeakyRateLimiterMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}
func (middleware *LeakyRateLimiterMiddleware) OnRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcRequest(session)
}
func (middleware *LeakyRateLimiterMiddleware) OnRpcResponse(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcResponse(session)
}

func (middleware *LeakyRateLimiterMiddleware) ProcessRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextProcessJSONRpcRequest(session)
}