package plugin

import (
	"github.com/zoowii/jsonrpc_proxygo/utils"
)

var log = utils.GetLogger("middleware_loader")

//func (chain *MiddlewareChain) LoadDisablePluginConfig(configInfo *config.ServerConfig) {
//	disablePluginConf := configInfo.Plugins.Disable
//	if disablePluginConf.Start && len(disablePluginConf.DisabledRpcMethods) > 0 {
//		disableMiddleware := disable.NewDisableMiddleware()
//		for _, item := range disablePluginConf.DisabledRpcMethods {
//			disableMiddleware.AddRpcMethodToBlacklist(item)
//		}
//		chain.InsertHead(disableMiddleware)
//	}
//}
//
//func (chain *MiddlewareChain) LoadCachePluginConfig(configInfo *config.ServerConfig) {
//	cachePluginConf := configInfo.Plugins.Caches
//	if len(cachePluginConf) > 0 {
//		cacheMiddleware := cache.NewCacheMiddleware()
//		usingCacheItemsCount := 0
//		for _, itemConf := range cachePluginConf {
//			if itemConf.ExpireSeconds <= 0 {
//				continue
//			}
//			methodNameForCache, jsonErr := cache.MakeMethodNameForCache(itemConf.Name, itemConf.ParamsForCache)
//			if jsonErr != nil {
//				log.Fatalln("parse cache params error", jsonErr)
//				return
//			}
//			item := &cache.CacheConfigItem{
//				MethodName:    methodNameForCache,
//				CacheDuration: time.Duration(itemConf.ExpireSeconds) * time.Second,
//			}
//			cacheMiddleware.AddCacheConfigItem(item)
//			usingCacheItemsCount++
//		}
//		if usingCacheItemsCount > 0 {
//			chain.InsertHead(cacheMiddleware)
//		}
//	}
//}
//
//func (chain *MiddlewareChain) LoadBeforeCachePluginConfig(configInfo *config.ServerConfig) {
//	beforeCachePluginConf := configInfo.Plugins.BeforeCacheConfigs
//	if len(beforeCachePluginConf) > 0 {
//		beforeCacheMiddleware := cache.NewBeforeCacheMiddleware()
//		usingBeforeCacheItemCount := 0
//		for _, itemConf := range beforeCachePluginConf {
//			if itemConf.FetchCacheKeyFromParamsCount <= 0 {
//				continue
//			}
//			item := &cache.BeforeCacheConfigItem{
//				MethodName:                   itemConf.MethodName,
//				FetchCacheKeyFromParamsCount: itemConf.FetchCacheKeyFromParamsCount,
//			}
//			beforeCacheMiddleware.AddConfigItem(item)
//			usingBeforeCacheItemCount++
//		}
//		if usingBeforeCacheItemCount > 0 {
//			chain.InsertHead(beforeCacheMiddleware)
//		}
//	}
//}

//func (chain *MiddlewareChain) LoadRateLimitPluginConfig(configInfo *config.ServerConfig) {
//	rateLimiterPluginConf := configInfo.Plugins.RateLimit
//	if rateLimiterPluginConf.Start {
//		if rateLimiterPluginConf.ConnectionRate <= 0 {
//			rateLimiterPluginConf.ConnectionRate = 1000000
//		}
//		if rateLimiterPluginConf.RpcRate <= 0 {
//			rateLimiterPluginConf.RpcRate = 10000000
//		}
//		rateLimiterMiddleware := rate_limit.NewRateLimiterMiddleware(rateLimiterPluginConf.ConnectionRate, rateLimiterPluginConf.RpcRate)
//		chain.InsertHead(rateLimiterMiddleware)
//	}
//}

//func (chain *MiddlewareChain) LoadUpstreamPluginConfig(configInfo *config.ServerConfig) {
//	upstreamPluginConf := configInfo.Plugins.Upstream
//	if len(upstreamPluginConf.TargetEndpoints) < 1 {
//		log.Fatalln("empty upstream target endpoints in config")
//		return
//	}
//	targetEndpoint := upstreamPluginConf.TargetEndpoints[0]
//	upstreamMiddleware := ws_upstream.NewWsUpstreamMiddleware(targetEndpoint.Url)
//	chain.InsertHead(upstreamMiddleware)
//}

//func (chain *MiddlewareChain) LoadLoadBalancePluginConfig(configInfo *config.ServerConfig) {
//	upstreamPluginConf := configInfo.Plugins.Upstream
//	if len(upstreamPluginConf.TargetEndpoints) <= 1 {
//		return
//	}
//	loadBalanceMiddleware := load_balancer.NewLoadBalanceMiddleware()
//	for _, itemConf := range upstreamPluginConf.TargetEndpoints {
//		if itemConf.Weight <= 0 {
//			log.Fatalln("invalid upstream weight", itemConf.Weight)
//			return
//		}
//		_, err := url.ParseRequestURI(itemConf.Url)
//		if err != nil {
//			log.Fatalln("invalid upstream target endpoint", itemConf.Url)
//			return
//		}
//		loadBalanceMiddleware.AddUpstreamItem(load_balancer.NewUpstreamItem(itemConf.Url, itemConf.Weight))
//	}
//	chain.InsertHead(loadBalanceMiddleware)
//}
//
//func (chain *MiddlewareChain) LoadStatisticPluginConfig(configInfo *config.ServerConfig) {
//	statisticPluginConf := configInfo.Plugins.Statistic
//	if statisticPluginConf.Start {
//		statisticMiddleware := statistic.NewStatisticMiddleware()
//		chain.InsertHead(statisticMiddleware)
//	}
//}
