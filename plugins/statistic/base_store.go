package statistic

import (
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"sync"
	"sync/atomic"
	"time"
)

type BaseMetricStore struct {
	MetricStore
	globalRpcMethodsCount *utils.MemoryCache
	globalRpcCallCount    uint64

	hourlyLock            sync.RWMutex
	hourlyStartTime       time.Time
	hourlyRpcMethodsCount *utils.MemoryCache
	hourlyRpcCallCount    uint64
}

func (store *BaseMetricStore) Init() error {
	store.globalRpcMethodsCount = utils.NewMemoryCache()
	store.globalRpcCallCount = 0
	store.hourlyStartTime = time.Now()
	store.hourlyRpcMethodsCount = utils.NewMemoryCache()
	store.hourlyRpcCallCount = 0
	return nil
}

func (store *BaseMetricStore) DumpStatInfo() (dump *StatData, err error) {
	dump = NewStatData()
	// global
	globalItems, err := store.globalRpcMethodsCount.DumpItems()
	if err != nil {
		return
	}
	for k, v := range globalItems {
		objectInt, objectErr := v.ObjectAsInt64()
		if objectErr != nil {
			err = objectErr
			return
		}
		dump.GlobalStat[k] = &MethodCallCacheInfo{
			Expiration: v.Expiration,
			CallCount:  objectInt,
		}
	}
	dump.GlobalRpcCallCount = store.globalRpcCallCount
	// hourly
	hourlyItems, err := store.hourlyRpcMethodsCount.DumpItems()
	if err != nil {
		return
	}
	for k, v := range hourlyItems {
		objectInt, objectErr := v.ObjectAsInt64()
		if objectErr != nil {
			err = objectErr
			return
		}
		dump.HourlyStat[k] = &MethodCallCacheInfo{
			Expiration: v.Expiration,
			CallCount:  objectInt,
		}
	}
	dump.HourlyRpcCallCount = store.hourlyRpcCallCount
	return
}

func (store *BaseMetricStore) incrementGlobalRpcMethodCalledCount(methodName string) {
	_, ok := store.globalRpcMethodsCount.Get(methodName)
	if ok {
		_ = store.globalRpcMethodsCount.Increment(methodName, 1)
	} else {
		store.globalRpcMethodsCount.SetDefault(methodName, 1)
	}
	for {
		newValue := store.globalRpcCallCount + 1
		if atomic.AddUint64(&store.globalRpcCallCount, 1) == newValue {
			break
		}
	}
}

func (store *BaseMetricStore) incrementHourlyRpcMethodCalledCount(methodName string) {
	store.hourlyLock.Lock()
	defer store.hourlyLock.Unlock()
	now := time.Now()
	if now.Sub(store.hourlyStartTime) > 1*time.Hour {
		store.hourlyStartTime = now
		store.hourlyRpcMethodsCount.Flush() // delete all items
		store.hourlyRpcCallCount = 0
	}
	_, ok := store.hourlyRpcMethodsCount.Get(methodName)
	if ok {
		_ = store.hourlyRpcMethodsCount.Increment(methodName, 1)
	} else {
		store.hourlyRpcMethodsCount.SetDefault(methodName, 1)
	}
	for {
		newValue := store.hourlyRpcCallCount + 1
		if atomic.AddUint64(&store.hourlyRpcCallCount, 1) == newValue {
			break
		}
	}
}

func (store *BaseMetricStore) addRpcMethodCall(methodName string) {
	store.incrementGlobalRpcMethodCalledCount(methodName)
	store.incrementHourlyRpcMethodCalledCount(methodName)
}
