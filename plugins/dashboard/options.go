package dashboard

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/registry"
)

type dashboardOptions struct {
	Endpoint string
	Context  context.Context
	Registry registry.Registry
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
