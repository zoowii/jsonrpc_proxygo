package dashboard

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
)

func LoadDashboardPluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	dashboardConfig := configInfo.Plugins.Dashboard
	if !dashboardConfig.Start {
		return
	}
	if len(dashboardConfig.Endpoint) < 1 {
		return
	}
	plugin := NewDashboardMiddleware(Endpoint(dashboardConfig.Endpoint))
	chain.InsertHead(plugin)
}
