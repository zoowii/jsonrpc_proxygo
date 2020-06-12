package statistic

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/registry"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"time"
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

func (store *dummyMetricStore) LogServiceDown(ctx context.Context, service *registry.Service) {

}

func (store *dummyMetricStore) QueryServiceDownLogs(ctx context.Context, offset int, limit int) (*ServiceLogListVo, error) {
	list := &ServiceLogListVo{
		Items: make([]*ServiceLogVo, 0),
		Total: 0,
	}
	return list, nil
}

func (store *dummyMetricStore) UpdateServiceHostPing(ctx context.Context, service *registry.Service, rtt time.Duration, connected bool) {

}

func (store *dummyMetricStore) QueryServiceHealthByUrl(ctx context.Context, service *registry.Service) (*ServiceHealthVo, error) {
	return nil, nil
}

func (store *dummyMetricStore) Name() string {
	return "dummy"
}

func newDefaultMetricStore() MetricStore {
	return &dummyMetricStore{}
}
