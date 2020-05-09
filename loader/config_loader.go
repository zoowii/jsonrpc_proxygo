package loader

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugins/cache"
	"github.com/zoowii/jsonrpc_proxygo/plugins/dashboard"
	"github.com/zoowii/jsonrpc_proxygo/plugins/disable"
	"github.com/zoowii/jsonrpc_proxygo/plugins/http_upstream"
	"github.com/zoowii/jsonrpc_proxygo/plugins/load_balancer"
	"github.com/zoowii/jsonrpc_proxygo/plugins/rate_limit"
	"github.com/zoowii/jsonrpc_proxygo/plugins/statistic"
	"github.com/zoowii/jsonrpc_proxygo/plugins/ws_upstream"
	"github.com/zoowii/jsonrpc_proxygo/providers"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"io/ioutil"
	"time"
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

	return
}

// 主动上报具体的服务状态给consul
func uploadConsulServiceHealthCheck(consulConfig *config.ConsulConfig, wholeConfig *config.ServerConfig) (err error) {
	err = utils.ConsulSubmitHealthChecker(consulConfig)
	return
}

func LoadConsulConfig(configInfo *config.ServerConfig) (err error) {
	if configInfo.Resolver != nil && configInfo.Resolver.Start {
		consulResolver := configInfo.Resolver
		configFileResolver := consulResolver.ConfigFileResolver
		if config.IsConsulResolver(configFileResolver) {
			log.Infof("loading config file from consul ", consulResolver.ConfigFileResolver)
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
			// resolver等属性只从本地读取，所以要修改从网络中获取的配置文件对象
			newConfigInfo.Resolver = configInfo.Resolver
			*configInfo = *newConfigInfo
			log.Infof("loaded new config from consul %s", configValue)
		}
		// 把本服务注册到consul agent
		if len(consulResolver.Endpoint) > 0 {
			err = utils.ConsulRegisterService(consulResolver, configInfo)
			if err != nil {
				return
			}
			// 后台开启心跳服务，自我检测服务状态上报
			healthCheckerIntervalSeconds := consulResolver.HealthCheckIntervalSeconds
			if healthCheckerIntervalSeconds <= 0 {
				healthCheckerIntervalSeconds = 60
			}
			go func() {
				ctx := context.Background()
				// 先立刻上报一次心跳
				err = uploadConsulServiceHealthCheck(consulResolver, configInfo)
				if err != nil {
					log.Errorf("uploadConsulServiceHealthCheck error %s", err.Error())
				}
				// 隔一段时间就发一次心跳包
				for {
					select {
					case <-ctx.Done():
						{
							break
						}
					case <-time.After(time.Duration(healthCheckerIntervalSeconds) * time.Second):
						{
							err = uploadConsulServiceHealthCheck(consulResolver, configInfo)
							if err != nil {
								log.Errorf("uploadConsulServiceHealthCheck error %s", err.Error())
							}
						}
					}
				}
			}()
		}
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
	http_upstream.LoadHttpUpstreamPluginConfig(server.MiddlewareChain, configInfo)
	load_balancer.LoadLoadBalancePluginConfig(server.MiddlewareChain, configInfo)
	disable.LoadDisablePluginConfig(server.MiddlewareChain, configInfo)
	cache.LoadCachePluginConfig(server.MiddlewareChain, configInfo)
	cache.LoadBeforeCachePluginConfig(server.MiddlewareChain, configInfo)
	rate_limit.LoadRateLimitPluginConfig(server.MiddlewareChain, configInfo)
	statistic.LoadStatisticPluginConfig(server.MiddlewareChain, configInfo)
	dashboard.LoadDashboardPluginConfig(server.MiddlewareChain, configInfo)
}
