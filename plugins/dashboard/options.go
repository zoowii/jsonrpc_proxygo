package dashboard

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/plugins/statistic"
	"github.com/zoowii/jsonrpc_proxygo/registry"
)

type dashboardOptions struct {
	Endpoint string
	Context  context.Context
	Registry registry.Registry
	Store statistic.MetricStore
}

func newDashBoardOptions() *dashboardOptions {
	return &dashboardOptions{
		Context: context.Background(),
	}
}

func Endpoint(endpoint string) common.Option {
	return func(options common.Options) {
		mOptions := options.(*dashboardOptions)
		mOptions.Endpoint = endpoint
	}
}

func WithContext(ctx context.Context) common.Option {
	return func(options common.Options) {
		mOptions := options.(*dashboardOptions)
		mOptions.Context = ctx
	}
}

func WithRegistry(r registry.Registry) common.Option {
	return func(options common.Options) {
		mOptions := options.(*dashboardOptions)
		mOptions.Registry = r
	}
}

func WithStore(store statistic.MetricStore) common.Option {
	return func(options common.Options) {
		mOptions := options.(*dashboardOptions)
		mOptions.Store = store
	}
}
