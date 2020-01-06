package rate_limit

import (
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"time"
)

type RateLimiterMiddleware struct {
	proxy.MiddlewareAdapter
	connLimiter Limiter
}

// NewRateLimiterMiddleware create rate-connLimiter middleware
// @param rate per second
func NewRateLimiterMiddleware(connectionRate, rpcRate int) *RateLimiterMiddleware {
	var limiter = NewTokenBucketLimiter(connectionRate, time.Second)
	return &RateLimiterMiddleware{
		connLimiter: limiter,
	}
}

func (middleware *RateLimiterMiddleware) Name() string {
	return "rate_limiter"
}

func (middleware *RateLimiterMiddleware) OnStart() (err error) {
	return middleware.NextOnStart()
}

func (middleware *RateLimiterMiddleware) OnConnection(session *proxy.ConnectionSession) (err error) {
	taken := middleware.connLimiter.Take()
	if !taken {
		err = errors.New("rate limit exceed")
		return
	}
	return middleware.NextOnConnection(session)
}

func (middleware *RateLimiterMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *RateLimiterMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}
func (middleware *RateLimiterMiddleware) OnRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcRequest(session)
}
func (middleware *RateLimiterMiddleware) OnRpcResponse(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcResponse(session)
}

func (middleware *RateLimiterMiddleware) ProcessRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextProcessJSONRpcRequest(session)
}