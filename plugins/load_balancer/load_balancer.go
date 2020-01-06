package load_balancer

import (
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"github.com/zoowii/jsonrpc_proxygo/utils"
)

var log = utils.GetLogger("load_balancer")

type UpstreamItem struct {
	Id int64
	TargetEndpoint string
	Weight int64
}

var upstreamItemIdGen int64 = 0

func NewUpstreamItem(targetEndpoint string, weight int64) *UpstreamItem {
	defer func() {
		upstreamItemIdGen++
	}()
	return &UpstreamItem{
		Id: upstreamItemIdGen,
		TargetEndpoint: targetEndpoint,
		Weight: weight,
	}
}

type LoadBalanceMiddleware struct {
	selector *WrrSelector
	UpstreamItems []*UpstreamItem
}

func NewLoadBalanceMiddleware() *LoadBalanceMiddleware {
	return &LoadBalanceMiddleware{
		selector: NewWrrSelector(),
	}
}

func (middleware *LoadBalanceMiddleware) AddUpstreamItem(item *UpstreamItem) *LoadBalanceMiddleware {
	middleware.UpstreamItems = append(middleware.UpstreamItems, item)
	middleware.selector.AddNode(item.Weight, item)
	return middleware
}

func (middleware *LoadBalanceMiddleware) Name() string {
	return "load_balance"
}

func (middleware *LoadBalanceMiddleware) selectTargetByWeight() *UpstreamItem {
	selected, err := middleware.selector.Next()
	if err != nil {
		log.Fatalln("load balance selector next error", err)
		return nil
	}
	selectedUpStreamItem, ok := selected.(*UpstreamItem)
	if !ok {
		return nil
	}
	return selectedUpStreamItem
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
	log.Debugf("selected upstream target item id#%d endpoint: %s\n",selectedTargetItem.Id, selectedTargetItem.TargetEndpoint)
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