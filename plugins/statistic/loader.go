package statistic

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
)

func LoadStatisticPluginConfig(chain *plugin.MiddlewareChain, configInfo *config.ServerConfig) {
	statisticPluginConf := configInfo.Plugins.Statistic
	if statisticPluginConf.Start {
		statisticMiddleware := NewStatisticMiddleware()
		chain.InsertHead(statisticMiddleware)
	}
}
