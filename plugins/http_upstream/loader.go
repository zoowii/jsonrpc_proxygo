package http_upstream

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
)

func LoadHttpUpstreamPluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	httpUpstreamConf := configInfo.Plugins.HttpUpstream
	if !httpUpstreamConf.Start {
		return
	}
	upstreamPluginConf := configInfo.Plugins.Upstream
	if len(upstreamPluginConf.TargetEndpoints) < 1 {
		log.Fatalln("empty upstream target endpoints in config")
		return
	}
	targetEndpoint := upstreamPluginConf.TargetEndpoints[0]
	upstreamMiddleware, err := NewHttpUpstreamMiddleware(HttpDefaultTargetEndpoint(targetEndpoint.Url))
	if err != nil {
		log.Fatalln("load http upstream plugin config error", err)
	}
	chain.InsertHead(upstreamMiddleware)
}
