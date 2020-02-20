package loader

import (
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

var log = utils.GetLogger("loader")

func LoadConfigFromConfigJsonFile(configFilePath string) (configInfo *config.ServerConfig, err error) {
	configFileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return
	}
	configInfo = new(config.ServerConfig)
	err = config.UnmarshalServerConfigFromJson(configFileBytes, configInfo)
	if err != nil {
		return
	}
	if config.IsConsulResolver(configInfo.Resolver) {
		configFileResolver := configInfo.ConfigFileResolver
		if config.IsConsulResolver(configFileResolver) {
			log.Infof("loading config file from consul ", configInfo.ConfigFileResolver)
			// 需要尝试从consul kv http api加载配置文件
			var consulErr error
			configPair, consulErr := utils.ConsulGetKV(configFileResolver)
			if consulErr != nil {
				err = consulErr
				return
			}
			configValue := configPair.Value
			log.Infof("config value from consul: %s", configValue)

			newConfigInfo := &config.ServerConfig{}
			err = config.UnmarshalServerConfigFromJson([]byte(configValue), newConfigInfo)
			if err != nil {
				return
			}
			// resolver和config_file_resolver等属性只从本地读取，所以要修改从网络中获取的配置文件对象
			newConfigInfo.Resolver = configInfo.Resolver
			newConfigInfo.ConfigFileResolver = configInfo.ConfigFileResolver
			*configInfo = *newConfigInfo
			log.Infof("loaded new config from consul %s", configValue)
		}
		// TODO: 把本服务注册到consul agent
	}
	return
}

func SetLoggerFromConfig(configInfo *config.ServerConfig) {
	utils.SetLogLevel(configInfo.Log.Level)
	println("logger level set to " + configInfo.Log.Level)
	if len(configInfo.Log.OutputFile) > 0 {
		utils.AddFileOutputToLog(configInfo.Log.OutputFile)
		println("logger file to " + configInfo.Log.OutputFile)
	}
}

func LoadProviderFromConfig(configInfo *config.ServerConfig) providers.RpcProvider {
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
	return provider
}

func LoadPluginsFromConfig(server *proxy.ProxyServer, configInfo *config.ServerConfig) {
	ws_upstream.LoadWsUpstreamPluginConfig(server.MiddlewareChain, configInfo)
	load_balancer.LoadLoadBalancePluginConfig(server.MiddlewareChain, configInfo)
	disable.LoadDisablePluginConfig(server.MiddlewareChain, configInfo)
	cache.LoadCachePluginConfig(server.MiddlewareChain, configInfo)
	cache.LoadBeforeCachePluginConfig(server.MiddlewareChain, configInfo)
	rate_limit.LoadRateLimitPluginConfig(server.MiddlewareChain, configInfo)
	statistic.LoadStatisticPluginConfig(server.MiddlewareChain, configInfo)
}
