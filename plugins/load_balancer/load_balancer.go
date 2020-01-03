package load_balancer

import (
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"github.com/zoowii/jsonrpc_proxygo/utils"
)

var log = utils.GetLogger("load_balancer")

var upstreamItemIdGen = 0

type UpstreamItem struct {
	id int
	TargetEndpoint string
	Weight int64
	currentWeight int64
}

func NewUpstreamItem(targetEndpoint string, weight int64) *UpstreamItem {
	upstreamItemIdGen++
	return &UpstreamItem{
		id: upstreamItemIdGen,
		TargetEndpoint: targetEndpoint,
		Weight: weight,
		currentWeight: weight,
	}
}

type LoadBalanceMiddleware struct {
	UpstreamItems []*UpstreamItem
}

func NewLoadBalanceMiddleware() *LoadBalanceMiddleware {
	return &LoadBalanceMiddleware{}
}

func (middleware *LoadBalanceMiddleware) AddUpstreamItem(item *UpstreamItem) *LoadBalanceMiddleware {
	middleware.UpstreamItems = append(middleware.UpstreamItems, item)
	return middleware
}

func (middleware *LoadBalanceMiddleware) Name() string {
	return "load_balance"
}

func (middleware *LoadBalanceMiddleware) selectTargetByWeight() *UpstreamItem {
	// use WeightedRound-Robin algorithm to select an target to use
	var totalWeight int64 = 0
	var maxWeight int64 = -1
	var maxWeightItem *UpstreamItem = nil
	for _, item := range middleware.UpstreamItems {
		totalWeight += item.Weight
		item.currentWeight += item.Weight
		if maxWeightItem == nil {
			maxWeight = item.currentWeight
			maxWeightItem = item
		} else if item.currentWeight > maxWeight {
			maxWeight = item.currentWeight
			maxWeightItem = item
		}
	}
	if maxWeightItem == nil {
		return nil
	}
	maxWeightItem.currentWeight -= totalWeight
	return maxWeightItem
}

func (middleware *LoadBalanceMiddleware) OnStart() (err error) {
	return
}

func (middleware *LoadBalanceMiddleware) OnConnection(session *proxy.ConnectionSession) (next bool, err error) {
	next = true

	selectedTargetItem := middleware.selectTargetByWeight()
	if selectedTargetItem == nil {
		err = errors.New("can't select one upstream target")
		return
	}
	log.Debugf("selected upstream target item id#%d endpoint: %s\n",selectedTargetItem.id, selectedTargetItem.TargetEndpoint)
	session.SelectedUpstreamTarget = &selectedTargetItem.TargetEndpoint
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