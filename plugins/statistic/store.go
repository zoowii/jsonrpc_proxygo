package statistic

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
)

type MetricStore interface {
	// LogRequest store request info. if {includeDebug}==true, store request's content and debug info
	LogRequest(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool)
	// LogRequest store response info. if {includeDebug}==true, store response's content and debug info
	logResponse(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool)
	Name() string
}

type dummyMetricStore struct {

}

func (store *dummyMetricStore) LogRequest(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool) {

}

func (store *dummyMetricStore) logResponse(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool) {

}

func (store *dummyMetricStore) Name() string {
	return "dummy"
}

func newDefaultMetricStore() MetricStore {
	return &dummyMetricStore{}
}
