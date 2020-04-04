package http_upstream

import (
	"github.com/zoowii/jsonrpc_proxygo/common"
	"time"
)

type httpUpstreamMiddlewareOptions struct {
	defaultTargetEndpoint string
	upstreamTimeout       time.Duration
}

func HttpDefaultTargetEndpoint(endpoint string) common.Option {
	return func(options common.Options) {
		mOptions := options.(*httpUpstreamMiddlewareOptions)
		mOptions.defaultTargetEndpoint = endpoint
	}
}

func HttpUpstreamTimeout(timeout time.Duration) common.Option {
	return func(options common.Options) {
		mOptions := options.(*httpUpstreamMiddlewareOptions)
		mOptions.upstreamTimeout = timeout
	}
}
