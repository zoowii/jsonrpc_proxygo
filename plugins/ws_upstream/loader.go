package ws_upstream

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
)

func LoadWsUpstreamPluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	httpUpstreamConf := configInfo.Plugins.HttpUpstream
	if httpUpstreamConf.Start {
		return
	}
	upstreamPluginConf := configInfo.Plugins.Upstream
	if len(upstreamPluginConf.TargetEndpoints) < 1 {
		log.Fatalln("empty upstream target endpoints in config")
		return
	}
	targetEndpoint := upstreamPluginConf.TargetEndpoints[0]
	upstreamMiddleware := NewWsUpstreamMiddleware(WsDefaultTargetEndpoint(targetEndpoint.Url))
	chain.InsertHead(upstreamMiddleware)
}
