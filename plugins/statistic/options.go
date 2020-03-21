package statistic

import (
	"github.com/zoowii/jsonrpc_proxygo/config"
)

type MetricOptions struct {
	store              MetricStore // metric store strategy
	dumpIntervalOpened bool        // whether dump metric status interval
}

func DbStore(dbUrl string) config.Option {
	return func(options config.Options) {
		mOptions, _ := options.(*MetricOptions)
		dbStore := newMetricDbStore(dbUrl)
		mOptions.store = dbStore
		err := dbStore.Init()
		if err != nil {
			log.Error("metric db init error ", err)
		}
	}
}

func DumpInterval() config.Option {
	return func(options config.Options) {
		mOptions, _ := options.(*MetricOptions)
		mOptions.dumpIntervalOpened = true
	}
}
