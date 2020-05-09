package statistic

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"sync"
	"time"
)

type MetricStore interface {
	Init() error
	// LogRequest store request info. if {includeDebug}==true, store request's content and debug info
	LogRequest(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool)
	// LogRequest store response info. if {includeDebug}==true, store response's content and debug info
	logResponse(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool)
	Name() string

	DumpGlobalStatInfoJson() (jsonBytes []byte, err error)
	DumpHourlyStatInfoJson() (jsonBytes []byte, err error)
	IncrementGlobalRpcMethodCalledCount(methodName string)
	IncrementHourlyRpcMethodCalledCount(methodName string)
}

type BaseMetricStore struct {
	MetricStore
	globalRpcMethodsCount *utils.MemoryCache

	hourlyLock            sync.RWMutex
	hourlyStartTime       time.Time
	hourlyRpcMethodsCount *utils.MemoryCache
}

func (store *BaseMetricStore) Init() error {
	store.globalRpcMethodsCount = utils.NewMemoryCache()
	store.hourlyStartTime = time.Now()
	store.hourlyRpcMethodsCount = utils.NewMemoryCache()
	return nil
}

func (store *BaseMetricStore) DumpGlobalStatInfoJson() (jsonBytes []byte, err error) {
	return store.globalRpcMethodsCount.DumpItems()
}

func (store *BaseMetricStore) DumpHourlyStatInfoJson() (jsonBytes []byte, err error) {
	return store.hourlyRpcMethodsCount.DumpItems()
}

func (store *BaseMetricStore) IncrementGlobalRpcMethodCalledCount(methodName string) {
	_, ok := store.globalRpcMethodsCount.Get(methodName)
	if ok {
		_ = store.globalRpcMethodsCount.Increment(methodName, 1)
	} else {
		store.globalRpcMethodsCount.SetDefault(methodName, 1)
	}
}

func (store *BaseMetricStore) IncrementHourlyRpcMethodCalledCount(methodName string) {
	store.hourlyLock.Lock()
	defer store.hourlyLock.Unlock()
	now := time.Now()
	if now.Sub(store.hourlyStartTime) > 1*time.Hour {
		store.hourlyStartTime = now
		store.hourlyRpcMethodsCount.Flush() // delete all items
	}
	_, ok := store.hourlyRpcMethodsCount.Get(methodName)
	if ok {
		_ = store.hourlyRpcMethodsCount.Increment(methodName, 1)
	} else {
		store.hourlyRpcMethodsCount.SetDefault(methodName, 1)
	}
}

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

func (store *dummyMetricStore) Name() string {
	return "dummy"
}

func newDefaultMetricStore() MetricStore {
	return &dummyMetricStore{}
}
