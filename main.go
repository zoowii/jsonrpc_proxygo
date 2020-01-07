package main

import (
	"flag"
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugins/cache"
	"github.com/zoowii/jsonrpc_proxygo/plugins/disable"
	"github.com/zoowii/jsonrpc_proxygo/plugins/load_balancer"
	"github.com/zoowii/jsonrpc_proxygo/plugins/rate_limit"
	"github.com/zoowii/jsonrpc_proxygo/plugins/statistic"
	"github.com/zoowii/jsonrpc_proxygo/plugins/ws_upstream"
	"github.com/zoowii/jsonrpc_proxygo/providers"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"io/ioutil"
)

func main() {
	utils.Init()
	var log = utils.GetLogger("main")

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

	utils.SetLogLevel(configInfo.Log.Level)
	if len(configInfo.Log.OutputFile) > 0 {
		utils.AddFileOutputToLog(configInfo.Log.OutputFile)
	}

	addr := configInfo.Endpoint
	log.Info("to start proxy server on " + addr)
	
	var provider providers.RpcProvider
	switch configInfo.Provider {
	case "http":
		provider = providers.NewHttpJsonRpcProvider(addr, "/", &providers.HttpJsonRpcProviderOptions{
			TimeoutSeconds: 30,
		})
	case "websocket":
		provider = providers.NewWebSocketJsonRpcProvider(addr, "/")
	default:
		provider = providers.NewWebSocketJsonRpcProvider(addr, "/")
	}
	
	server := proxy.NewProxyServer(provider)

	ws_upstream.LoadUpstreamPluginConfig(server.MiddlewareChain, &configInfo)
	load_balancer.LoadLoadBalancePluginConfig(server.MiddlewareChain, &configInfo)
	disable.LoadDisablePluginConfig(server.MiddlewareChain, &configInfo)
	cache.LoadCachePluginConfig(server.MiddlewareChain, &configInfo)
	cache.LoadBeforeCachePluginConfig(server.MiddlewareChain, &configInfo)
	rate_limit.LoadRateLimitPluginConfig(server.MiddlewareChain, &configInfo)
	statistic.LoadStatisticPluginConfig(server.MiddlewareChain, &configInfo)

	err = server.StartMiddlewares()
	if err != nil {
		log.Panic("start middlewares error", err.Error())
		return
	}
	log.Printf("loaded middlewares are(count %d):\n", len(server.MiddlewareChain.Middlewares))
	for _, middleware := range server.MiddlewareChain.Middlewares {
		log.Printf("\t- middleware %s\n", middleware.Name())
	}
	server.Start()
}
