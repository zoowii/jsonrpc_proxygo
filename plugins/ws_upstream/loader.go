package ws_upstream

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
)

func LoadUpstreamPluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	upstreamPluginConf := configInfo.Plugins.Upstream
	if len(upstreamPluginConf.TargetEndpoints) < 1 {
		log.Fatalln("empty upstream target endpoints in config")
		return
	}
	targetEndpoint := upstreamPluginConf.TargetEndpoints[0]
	upstreamMiddleware := NewWsUpstreamMiddleware(targetEndpoint.Url)
	chain.InsertHead(upstreamMiddleware)
}
