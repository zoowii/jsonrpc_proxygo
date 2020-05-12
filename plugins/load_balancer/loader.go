package load_balancer

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/registry"
	"net/url"
)

func LoadLoadBalancePluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig, r registry.Registry) {
	upstreamPluginConf := configInfo.Plugins.Upstream
	targetEndpoints := upstreamPluginConf.TargetEndpoints
	if len(targetEndpoints) <= 1 {
		return // 初始至少需要提供一个upstream target
	}
	loadBalanceMiddleware := NewLoadBalanceMiddleware()
	for _, itemConf := range targetEndpoints {
		if itemConf.Ignore {
			continue
		}
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

		if r != nil {
			// register service to registry
			err = r.RegisterService(&registry.Service{
				Name: "upstream",
				Url:  itemConf.Url,
			})
			if err != nil {
				log.Fatalln("register upstream to registry error", err)
				return
			}
		}
	}
	chain.InsertHead(loadBalanceMiddleware)
	// TODO: load balance watch registry event channel
}
