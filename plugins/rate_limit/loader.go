package rate_limit

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
)

func LoadRateLimitPluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	rateLimiterPluginConf := configInfo.Plugins.RateLimit
	if rateLimiterPluginConf.Start {
		if rateLimiterPluginConf.ConnectionRate <= 0 {
			rateLimiterPluginConf.ConnectionRate = 1000000
		}
		if rateLimiterPluginConf.RpcRate <= 0 {
			rateLimiterPluginConf.RpcRate = 10000000
		}
		rateLimiterMiddleware := NewRateLimiterMiddleware(rateLimiterPluginConf.ConnectionRate, rateLimiterPluginConf.RpcRate)
		chain.InsertHead(rateLimiterMiddleware)
	}
}
