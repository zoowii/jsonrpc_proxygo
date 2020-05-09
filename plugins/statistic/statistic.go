package statistic

/**
 * statistic middleware
 * async calculate statistic metrics and publish to admin users
 * TODO: publish metrics. now just dump metrics to log interval
 */

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"time"
)

var log = utils.GetLogger("statistic")

// 使用中的metricStore，用来暴露给其他组件访问
var UsedMetricStore MetricStore = nil

type StatisticMiddleware struct {
	plugin.MiddlewareAdapter
	rpcRequestsReceived  chan *rpc.JSONRpcRequestSession
	rpcResponsesReceived chan *rpc.JSONRpcRequestSession
	//
	//globalRpcMethodsCount *utils.MemoryCache
	//
	//hourlyLock            sync.RWMutex
	//hourlyStartTime       time.Time
	//hourlyRpcMethodsCount *utils.MemoryCache

	metricOptions *MetricOptions
	store MetricStore
}

func NewStatisticMiddleware(options ...common.Option) *StatisticMiddleware {
	const maxRpcChannelSize = 10000

	mOptions := &MetricOptions{}
	for _, o := range options {
		o(mOptions)
	}

	var store MetricStore
	if mOptions.store != nil {
		store = mOptions.store
	} else {
		store = newDefaultMetricStore()
	}

	UsedMetricStore = store

	err := store.Init()
	if err != nil {
		log.Fatalf("statistic store init error %s", err.Error())
	}

	return &StatisticMiddleware{
		rpcRequestsReceived:   make(chan *rpc.JSONRpcRequestSession, maxRpcChannelSize),
		rpcResponsesReceived:  make(chan *rpc.JSONRpcRequestSession, maxRpcChannelSize),
		//globalRpcMethodsCount: utils.NewMemoryCache(),
		//hourlyStartTime:       time.Now(),
		//hourlyRpcMethodsCount: utils.NewMemoryCache(),
		metricOptions:         mOptions,
		store:                 store,
	}
}

func (middleware *StatisticMiddleware) Name() string {
	return "statistic"
}

func getMethodNameForRpcStatistic(session *rpc.JSONRpcRequestSession) string {
	if session.MethodNameForCache != nil {
		return *session.MethodNameForCache // cache name is more acurrate for statistic
	}
	return session.Request.Method
}

func (middleware *StatisticMiddleware) OnStart() (err error) {
	go func() {
		ctx := context.Background()

		store := middleware.store

		dumpIntervalOpened := middleware.metricOptions.dumpIntervalOpened
		dumpTick := time.Tick(60 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-dumpTick:
				if !dumpIntervalOpened {
					continue
				}
				// notify user every some time
				log.Info("start dump statistic info")
				globalStatJson, err := middleware.store.DumpGlobalStatInfoJson()
				if err != nil {
					log.Error("dump globalRpcMethodsCount error", err)
					continue
				}
				hourlyStatJson, err := middleware.store.DumpHourlyStatInfoJson()
				if err != nil {
					log.Error("dump hourlyRpcMethodsCount error", err)
					continue
				}
				log.Infof("globalRpcMethodsCount: %s", string(globalStatJson))
				log.Infof("hourlyRpcMethodsCount: %s", string(hourlyStatJson))
			case reqSession := <-middleware.rpcRequestsReceived:
				methodNameForStatistic := getMethodNameForRpcStatistic(reqSession)

				// update global rpc methods called count
				store.IncrementGlobalRpcMethodCalledCount(methodNameForStatistic)

				// update hourly rpc methods called count
				store.IncrementHourlyRpcMethodCalledCount(methodNameForStatistic)

				// TODO: 根据策略随机采样或者全部记录请求和返回的数据
				includeDebug := true
				store.LogRequest(ctx, reqSession, includeDebug)
			case resSession := <-middleware.rpcResponsesReceived:
				includeDebug := true
				store.logResponse(ctx, resSession, includeDebug)
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
	return middleware.NextOnStart()
}

func (middleware *StatisticMiddleware) OnConnection(session *rpc.ConnectionSession) (err error) {
	return middleware.NextOnConnection(session)
}

func (middleware *StatisticMiddleware) OnConnectionClosed(session *rpc.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *StatisticMiddleware) OnWebSocketFrame(session *rpc.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}
func (middleware *StatisticMiddleware) OnRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	middleware.rpcRequestsReceived <- session
	return middleware.NextOnJSONRpcRequest(session)
}
func (middleware *StatisticMiddleware) OnRpcResponse(session *rpc.JSONRpcRequestSession) (err error) {
	err = middleware.NextOnJSONRpcResponse(session)
	middleware.rpcResponsesReceived <- session
	return
}

func (middleware *StatisticMiddleware) ProcessRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	return middleware.NextProcessJSONRpcRequest(session)
}
