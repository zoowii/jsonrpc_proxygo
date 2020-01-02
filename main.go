package main

import (
	"flag"
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugins/cache"
	"github.com/zoowii/jsonrpc_proxygo/plugins/ws_upstream"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	// TODO: load config from yml config file
	configPath := flag.String("config", "server.json", "configuration file path(default server.json)")
	flag.Parse()
	configFileBytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalln(err)
		return
	}
	var configInfo config.ServerConfig
	err = config.UnmarshalServerConfigFromJson(configFileBytes, &configInfo)
	if err != nil {
		log.Fatalln(err)
		return
	}
	addr := configInfo.Endpoint
	log.Println("to start proxy server on " + addr)
	server := proxy.NewProxyServer(addr)
	upstreamPluginConf := configInfo.Plugins.Upstream
	if len(upstreamPluginConf.TargetEndpoints) < 1 {
		log.Fatalln("empty upstream target endpoints in config")
		return
	}
	targetEndpoint := upstreamPluginConf.TargetEndpoints[0]
	server.MiddlewareChain.InsertHead(
		ws_upstream.NewWsUpstreamMiddleware(targetEndpoint),
	)
	cachePluginConf := configInfo.Plugins.Caches
	if len(cachePluginConf) > 0 {
		cacheMiddleware := cache.NewCacheMiddleware()
		usingCacheItemsCount := 0
		for _, itemConf := range cachePluginConf {
			if itemConf.ExpireSeconds <= 0 {
				continue
			}
			item := &cache.CacheConfigItem{
				MethodName:    itemConf.Name,
				CacheDuration: time.Duration(itemConf.ExpireSeconds) * time.Second,
			}
			cacheMiddleware.AddCacheConfigItem(item)
			usingCacheItemsCount++
		}
		if usingCacheItemsCount > 0 {
			server.MiddlewareChain.InsertHead(cacheMiddleware)
		}
	}
	beforeCachePluginConf := configInfo.Plugins.BeforeCacheConfigs
	if len(beforeCachePluginConf) > 0 {
		beforeCacheMiddleware := cache.NewBeforeCacheMiddleware()
		usingBeforeCacheItemCount := 0
		for _, itemConf := range beforeCachePluginConf {
			if itemConf.FetchCacheKeyFromParamsCount <= 0 {
				continue
			}
			item := &cache.BeforeCacheConfigItem{
				MethodName:                   itemConf.MethodName,
				FetchCacheKeyFromParamsCount: itemConf.FetchCacheKeyFromParamsCount,
			}
			beforeCacheMiddleware.AddConfigItem(item)
			usingBeforeCacheItemCount++
		}
		if usingBeforeCacheItemCount > 0 {
			server.MiddlewareChain.InsertHead(beforeCacheMiddleware)
		}
	}
	log.Printf("loaded middlewares are(count %d):\n", len(server.MiddlewareChain.Middlewares))
	for _, middleware := range server.MiddlewareChain.Middlewares {
		log.Printf("\t- middleware %s\n", middleware.Name())
	}
	server.Start()
}
