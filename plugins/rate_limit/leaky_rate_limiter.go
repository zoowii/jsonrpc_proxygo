package rate_limit

// TODO: 漏桶算法的限流中间件

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

type LeakyRateLimiterMiddleware struct {

}

func (middleware *LeakyRateLimiterMiddleware) Name() string {
	return "leaky_rate_limiter"
}

func (middleware *LeakyRateLimiterMiddleware) OnStart() (err error) {
	return
}

func (middleware *LeakyRateLimiterMiddleware) OnConnection(session *proxy.ConnectionSession) (next bool, err error) {
	next = true
	return
}

func (middleware *LeakyRateLimiterMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (next bool, err error) {
	next = true
	return
}

func (middleware *LeakyRateLimiterMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (next bool, err error) {
	next = true
	return
}
func (middleware *LeakyRateLimiterMiddleware) OnJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}
func (middleware *LeakyRateLimiterMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}

func (middleware *LeakyRateLimiterMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}