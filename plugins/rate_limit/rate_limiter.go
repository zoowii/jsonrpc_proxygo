package rate_limit

import (
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"time"
)

type RateLimiterMiddleware struct {
	plugin.MiddlewareAdapter
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

func (middleware *RateLimiterMiddleware) OnConnection(session *rpc.ConnectionSession) (err error) {
	taken := middleware.connLimiter.Take()
	if !taken {
		err = errors.New("rate limit exceed")
		return
	}
	return middleware.NextOnConnection(session)
}

func (middleware *RateLimiterMiddleware) OnConnectionClosed(session *rpc.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *RateLimiterMiddleware) OnWebSocketFrame(session *rpc.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}
func (middleware *RateLimiterMiddleware) OnRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcRequest(session)
}
func (middleware *RateLimiterMiddleware) OnRpcResponse(session *rpc.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcResponse(session)
}

func (middleware *RateLimiterMiddleware) ProcessRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	return middleware.NextProcessJSONRpcRequest(session)
}
