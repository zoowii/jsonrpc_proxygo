package cache

import (
	"encoding/json"
	"fmt"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"time"
)

var log = utils.GetLogger("cache")

type CacheConfigItem struct {
	MethodName string
	CacheDuration time.Duration
}

type CacheMiddleware struct {
	cacheConfigItems []*CacheConfigItem
	cacheConfigItemsMap map[string]*CacheConfigItem // methodName => *CacheConfigItem

	rpcCache *utils.MemoryCache
}

func NewCacheMiddleware(cacheConfigItems ...*CacheConfigItem) *CacheMiddleware {
	cacheConfigItemsMap := make(map[string]*CacheConfigItem)
	result := &CacheMiddleware{
		cacheConfigItems: nil,
		cacheConfigItemsMap: cacheConfigItemsMap,
		rpcCache: utils.NewMemoryCache(),
	}
	for _, item := range cacheConfigItems {
		_ = result.AddCacheConfigItem(item)
	}
	return result
}

func (middleware *CacheMiddleware) AddCacheConfigItem(item *CacheConfigItem) *CacheMiddleware {
	middleware.cacheConfigItems = append(middleware.cacheConfigItems, item)
	if item != nil {
		middleware.cacheConfigItemsMap[item.MethodName] = item
	}
	return middleware
}

func (middleware *CacheMiddleware) Name() string {
	return "cache"
}

func (middleware *CacheMiddleware) OnConnection(session *proxy.ConnectionSession) (bool, error) {
	return true, nil
}

func (middleware *CacheMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (bool, error) {
	return true, nil
}

func (middleware *CacheMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (bool, error) {
	return true, nil
}

func fetchRpcRequestParams(rpcRequest *proxy.JSONRpcRequest, fetchParamsCount int) (result []interface{}) {
	if fetchParamsCount < 1 || rpcRequest.Params == nil {
		return
	}
	paramsArray, ok := rpcRequest.Params.([]interface{})
	if !ok {
		return
	}
	if fetchParamsCount > len(paramsArray) {
		fetchParamsCount = len(paramsArray)
	}
	result = make([]interface{}, fetchParamsCount)
	for i:=0;i<fetchParamsCount;i++ {
		result[i] = paramsArray[i]
	}
	return
}

func (middleware *CacheMiddleware) getCacheConfigItem(session *proxy.JSONRpcRequestSession) (result *CacheConfigItem, ok bool) {
	methodNameForCache := middleware.getMethodNameForCache(session)
	result, ok = middleware.cacheConfigItemsMap[methodNameForCache]
	return
}

func (middleware *CacheMiddleware) getMethodNameForCache(session *proxy.JSONRpcRequestSession) string {
	if session.MethodNameForCache != nil {
		return *session.MethodNameForCache
	}
	return session.Request.Method
}

func (middleware *CacheMiddleware) OnJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true

	methodNameForCache := middleware.getMethodNameForCache(session)
	if _, ok := middleware.getCacheConfigItem(session); !ok {
		return
	}
	cacheKey, err := middleware.cacheKeyForRpcMethod(methodNameForCache, session.Request.Params)
	if err != nil {
		log.Fatalln("cache key for rpc method error", err)
		return
	}
	cached, ok := middleware.rpcCache.Get(cacheKey)
	if !ok {
		return
	}
	cachedItem, ok := cached.(*rpcResponseCacheItem)
	if !ok {
		return
	}
	// need replace cachedItem's rpc request id
	newRes, err := proxy.CloneJSONRpcResponse(cachedItem.response)
	if err != nil {
		return
	}
	newRes.Id = session.Request.Id
	newResBytes, err := json.Marshal(newRes)
	if err != nil {
		return
	}
	session.Response = newRes
	session.RequestBytes = newResBytes
	session.ResponseSetByCache = true
	next = false
	log.Debugf("rpc method-for-cache %s hit cache", methodNameForCache)
	return
}

// cache by methodName + allRpcParams
func (middleware *CacheMiddleware) cacheKeyForRpcMethod(rpcMethodName string, rpcParams interface{}) (result string, err error) {
	rpcParamsBytes, err := json.Marshal(rpcParams)
	if err != nil {
		return
	}
	result = fmt.Sprintf("cache_rpc_%s$%s", rpcMethodName, string(rpcParamsBytes))
	return
}

type rpcResponseCacheItem struct {
	response *proxy.JSONRpcResponse
	responseBytes []byte
}

func (middleware *CacheMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	if session.ResponseSetByCache {
		next = false
		// log.Debug("ResponseSetByCache set before")
		return
	}
	methodNameForCache := middleware.getMethodNameForCache(session)
	cacheConfigItem, ok := middleware.getCacheConfigItem(session)
	if !ok {
		return
	}
	rpcRes := session.Response
	rpcResBytes := session.RequestBytes
	cacheKey, err := middleware.cacheKeyForRpcMethod(methodNameForCache, session.Request.Params)
	if err != nil {
		return
	}
	middleware.rpcCache.Set(cacheKey, &rpcResponseCacheItem{
		response: rpcRes,
		responseBytes: rpcResBytes,
	}, cacheConfigItem.CacheDuration)
	log.Debugf("rpc method-for-cache %s cached\n", methodNameForCache)
	return
}

func (middleware *CacheMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, bool error) {
	next = true
	if session.ResponseSetByCache {
		next = false
	}
	return
}