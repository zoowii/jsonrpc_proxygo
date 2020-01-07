package cache

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"time"
)

func LoadCachePluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	cachePluginConf := configInfo.Plugins.Caches
	if len(cachePluginConf) > 0 {
		cacheMiddleware := NewCacheMiddleware()
		usingCacheItemsCount := 0
		for _, itemConf := range cachePluginConf {
			if itemConf.ExpireSeconds <= 0 {
				continue
			}
			methodNameForCache, jsonErr := MakeMethodNameForCache(itemConf.Name, itemConf.ParamsForCache)
			if jsonErr != nil {
				log.Fatalln("parse cache params error", jsonErr)
				return
			}
			item := &CacheConfigItem{
				MethodName:    methodNameForCache,
				CacheDuration: time.Duration(itemConf.ExpireSeconds) * time.Second,
			}
			cacheMiddleware.AddCacheConfigItem(item)
			usingCacheItemsCount++
		}
		if usingCacheItemsCount > 0 {
			chain.InsertHead(cacheMiddleware)
		}
	}
}

func LoadBeforeCachePluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	beforeCachePluginConf := configInfo.Plugins.BeforeCacheConfigs
	if len(beforeCachePluginConf) > 0 {
		beforeCacheMiddleware := NewBeforeCacheMiddleware()
		usingBeforeCacheItemCount := 0
		for _, itemConf := range beforeCachePluginConf {
			if itemConf.FetchCacheKeyFromParamsCount <= 0 {
				continue
			}
			item := &BeforeCacheConfigItem{
				MethodName:                   itemConf.MethodName,
				FetchCacheKeyFromParamsCount: itemConf.FetchCacheKeyFromParamsCount,
			}
			beforeCacheMiddleware.AddConfigItem(item)
			usingBeforeCacheItemCount++
		}
		if usingBeforeCacheItemCount > 0 {
			chain.InsertHead(beforeCacheMiddleware)
		}
	}
}

