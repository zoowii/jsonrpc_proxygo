package load_balancer

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"net/url"
)

func LoadLoadBalancePluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	upstreamPluginConf := configInfo.Plugins.Upstream
	if len(upstreamPluginConf.TargetEndpoints) <= 1 {
		return
	}
	loadBalanceMiddleware := NewLoadBalanceMiddleware()
	for _, itemConf := range upstreamPluginConf.TargetEndpoints {
		if itemConf.Weight <= 0 {
			log.Fatalln("invalid upstream weight", itemConf.Weight)
			return
		}
		_, err := url.ParseRequestURI(itemConf.Url)
		if err != nil {
			log.Fatalln("invalid upstream target endpoint", itemConf.Url)
			return
		}
		loadBalanceMiddleware.AddUpstreamItem(NewUpstreamItem(itemConf.Url, itemConf.Weight))
	}
	chain.InsertHead(loadBalanceMiddleware)
}

