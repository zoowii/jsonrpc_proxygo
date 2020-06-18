package statistic

import (
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/registry"
)

func LoadStatisticPluginConfig(chain *plugin.MiddlewareChain,
	configInfo *config.ServerConfig, r registry.Registry) (result *StatisticMiddleware) {
	statisticPluginConf := configInfo.Plugins.Statistic
	if statisticPluginConf.Start {
		options := make([]common.Option, 0)
		if statisticPluginConf.Store.DumpIntervalOpened {
			options = append(options, DumpInterval())
			log.Info("statistic plugin load DumpInterval option")
		}
		if statisticPluginConf.Store.Type == metricDbStoreName && len(statisticPluginConf.Store.DbUrl) > 0 {
			options = append(options, DbStore(statisticPluginConf.Store.DbUrl))
			log.Info("statistic plugin load DbStore option")
		}
		options = append(options, SetRegistry(r))
		log.Info("statistic plugin load registry option")

		result = NewStatisticMiddleware(options...)
		chain.InsertHead(result)
	}
	return
}
