package dashboard

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/common"
)

type dashboardOptions struct {
	Endpoint string
	Context context.Context
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
