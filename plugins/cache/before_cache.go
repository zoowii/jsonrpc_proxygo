package cache

import (
	"encoding/json"

	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

type BeforeCacheConfigItem struct {
	MethodName                   string
	FetchCacheKeyFromParamsCount int /* eg. when rpc params: [2, "info", "hello"], and FetchCacheKeyFromParamsCount==2, then methodName for cache middleware will be "call$2$\"info\""*/
}

/**
 * a middleware inserted before cache middleware to fetch `methodName for cache` from rpc request
 * eg. when rpc format is {method: "callOrOther", params: ["realMethodName", ...otherArgs]} for some methods
 */
type BeforeCacheMiddleware struct {
	proxy.MiddlewareAdapter
	BeforeCacheConfigItems []*BeforeCacheConfigItem
}

func NewBeforeCacheMiddleware() *BeforeCacheMiddleware {
	return &BeforeCacheMiddleware{}
}

func (middleware *BeforeCacheMiddleware) AddConfigItem(item *BeforeCacheConfigItem) *BeforeCacheMiddleware {
	// TODO: 构造一个字典树，用来使用时快速定位一个请求是否需要做beforeCache的处理
	if item == nil || item.FetchCacheKeyFromParamsCount < 1 {
		return middleware
	}
	middleware.BeforeCacheConfigItems = append(middleware.BeforeCacheConfigItems, item)
	return middleware
}

func (middleware *BeforeCacheMiddleware) Name() string {
	return "before_cache"
}

func (middleware *BeforeCacheMiddleware) OnStart() (err error) {
	return
}

func (middleware *BeforeCacheMiddleware) OnConnection(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnection(session)
}

func (middleware *BeforeCacheMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *BeforeCacheMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}

func (middleware *BeforeCacheMiddleware) findBeforeCacheConfigItem(rpcReq *proxy.JSONRpcRequest) (result *BeforeCacheConfigItem, ok bool) {
	methodName := rpcReq.Method
	rpcParams := rpcReq.Params
	rpcParamsArray, parseArrayOk := rpcParams.([]interface{})
	if !parseArrayOk {
		return
	}
	for _, item := range middleware.BeforeCacheConfigItems {
		if item.MethodName != methodName {
			continue
		}
		if len(rpcParamsArray) < item.FetchCacheKeyFromParamsCount {
			continue
		}
		result = item
		ok = true
		return
	}
	return
}

func MakeMethodNameForCache(methodName string, paramsArray []interface{}) (result string, err error) {
	result = methodName
	for i := 0; i < len(paramsArray); i++ {
		result += "$"
		argBytes, jsonErr := json.Marshal(paramsArray[i])
		if jsonErr != nil {
			err = jsonErr
			return
		}
		result += string(argBytes)
	}
	return
}

func (middleware *BeforeCacheMiddleware) OnRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	defer func() {
		if err == nil {
			err = middleware.NextOnJSONRpcRequest(session)
		}
	}()
	rpcReq := session.Request
	beforeCacheConfigItem, ok := middleware.findBeforeCacheConfigItem(rpcReq)
	if !ok {
		return
	}
	rpcParams := rpcReq.Params
	rpcParamsArray, parseArrayOk := rpcParams.([]interface{})
	if !parseArrayOk {
		return
	}
	fetchCacheKeyFromParamsCount := beforeCacheConfigItem.FetchCacheKeyFromParamsCount
	methodNameForCache, jsonErr := MakeMethodNameForCache(rpcReq.Method, rpcParamsArray[0:fetchCacheKeyFromParamsCount])
	if jsonErr != nil {
		log.Println("[before-cache] before_cache middleware parse param json error:", jsonErr)
		return
	}
	session.MethodNameForCache = &methodNameForCache

	// log.Debugf("[before-cache] methodNameForCache %s set\n", methodNameForCache)
	return
}
func (middleware *BeforeCacheMiddleware) OnRpcResponse(session *proxy.JSONRpcRequestSession) (error) {
	return middleware.NextOnJSONRpcResponse(session)
}

func (middleware *BeforeCacheMiddleware) ProcessRpcRequest(session *proxy.JSONRpcRequestSession) (error) {
	return middleware.NextProcessJSONRpcRequest(session)
}
