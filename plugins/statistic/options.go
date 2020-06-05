package statistic

import (
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/registry"
)

type MetricOptions struct {
	store              MetricStore // metric store strategy
	r                  registry.Registry
	dumpIntervalOpened bool // whether dump metric status interval
}

func DbStore(dbUrl string) common.Option {
	return func(options common.Options) {
		mOptions, _ := options.(*MetricOptions)
		dbStore := newMetricDbStore(dbUrl)
		mOptions.store = dbStore
		err := dbStore.Init()
		if err != nil {
			log.Error("metric db init error ", err)
		}
	}
}

func DumpInterval() common.Option {
	return func(options common.Options) {
		mOptions, _ := options.(*MetricOptions)
		mOptions.dumpIntervalOpened = true
	}
}

func SetRegistry(r registry.Registry) common.Option {
	return func(options common.Options) {
		mOptions, _ := options.(*MetricOptions)
		mOptions.r = r
	}
}
