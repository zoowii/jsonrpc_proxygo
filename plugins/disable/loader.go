package disable

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
)

func LoadDisablePluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	disablePluginConf := configInfo.Plugins.Disable
	if disablePluginConf.Start && len(disablePluginConf.DisabledRpcMethods) > 0 {
		disableMiddleware := NewDisableMiddleware()
		for _, item := range disablePluginConf.DisabledRpcMethods {
			disableMiddleware.AddRpcMethodToBlacklist(item)
		}
		chain.InsertHead(disableMiddleware)
	}
}
