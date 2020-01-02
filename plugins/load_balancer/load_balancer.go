package load_balancer

// TODO

import (
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

type UpstreamItem struct {
	TargetEndpoint string
	Weight int64
}

func NewUpstreamItem(targetEndpoint string, weight int64) UpstreamItem {
	return UpstreamItem{
		TargetEndpoint: targetEndpoint,
		Weight: weight,
	}
}

type LoadBalanceMiddleware struct {
	UpstreamItems []UpstreamItem
}

func NewLoadBalanceMiddleware() *LoadBalanceMiddleware {
	return &LoadBalanceMiddleware{}
}

func (middleware *LoadBalanceMiddleware) AddUpstreamItem(item UpstreamItem) *LoadBalanceMiddleware {
	middleware.UpstreamItems = append(middleware.UpstreamItems, item)
	return middleware
}

func (middleware *LoadBalanceMiddleware) Name() string {
	return "load_balance"
}

func (middleware *LoadBalanceMiddleware) OnConnection(session *proxy.ConnectionSession) (next bool, err error) {
	next = true

	session.SelectedUpstreamTarget = &middleware.UpstreamItems[0].TargetEndpoint // now just use first upstream target

	// TODO: use WeightedRound-Robin algorithm to select an target to use in session
	return
}

func (middleware *LoadBalanceMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (next bool, err error) {
	next = true
	return
}

func (middleware *LoadBalanceMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (next bool, err error) {
	next = true
	return
}
func (middleware *LoadBalanceMiddleware) OnJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}
func (middleware *LoadBalanceMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}

func (middleware *LoadBalanceMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}