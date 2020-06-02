package statistic

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
)

type dummyMetricStore struct {
	BaseMetricStore
}

func (store *dummyMetricStore) Init() error {
	return store.BaseMetricStore.Init()
}

func (store *dummyMetricStore) LogRequest(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool) {

}

func (store *dummyMetricStore) logResponse(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool) {

}

func (store *dummyMetricStore) QueryRequestSpanList(ctx context.Context, form *QueryLogForm) (*RequestSpanListVo, error) {
	list := &RequestSpanListVo{
		Items: make([]*RequestSpanVo, 0),
		Total: 0,
	}
	return list, nil
}

func (store *dummyMetricStore) Name() string {
	return "dummy"
}

func newDefaultMetricStore() MetricStore {
	return &dummyMetricStore{}
}
