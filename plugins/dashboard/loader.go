package dashboard

import (
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/registry"
)

func LoadDashboardPluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig, r registry.Registry) {
	dashboardConfig := configInfo.Plugins.Dashboard
	if !dashboardConfig.Start {
		return
	}
	if len(dashboardConfig.Endpoint) < 1 {
		return
	}
	var registryOption common.Option
	if r != nil {
		registryOption = WithRegistry(r)
	}
	plugin := NewDashboardMiddleware(Endpoint(dashboardConfig.Endpoint), registryOption)
	chain.InsertHead(plugin)
}
