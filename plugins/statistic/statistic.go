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
	"github.com/zoowii/jsonrpc_proxygo/registry"
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

	metricOptions *MetricOptions
	store         MetricStore
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
		rpcRequestsReceived:  make(chan *rpc.JSONRpcRequestSession, maxRpcChannelSize),
		rpcResponsesReceived: make(chan *rpc.JSONRpcRequestSession, maxRpcChannelSize),
		metricOptions:        mOptions,
		store:                store,
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

		// 从registry监听服务上下限，如果有掉线，记录alert
		r := middleware.metricOptions.r // registry
		watcher, watcherErr := r.Watch()
		if watcherErr != nil {
			log.Error("watch registry error", watcherErr)
			return
		}
		registryEventChan := watcher.C()

		for {
			select {
			case <-ctx.Done():
				watcher.Close()
				return
			case <-dumpTick:
				if !dumpIntervalOpened {
					continue
				}
				// notify user every tick time
			case reqSession := <-middleware.rpcRequestsReceived:
				methodNameForStatistic := getMethodNameForRpcStatistic(reqSession)

				store.addRpcMethodCall(methodNameForStatistic)

				// TODO: 根据策略随机采样或者全部记录请求和返回的数据
				includeDebug := true
				store.LogRequest(ctx, reqSession, includeDebug)
			case resSession := <-middleware.rpcResponsesReceived:
				includeDebug := true
				store.logResponse(ctx, resSession, includeDebug)
			case registryEvent := <-registryEventChan:
				log.Infof("receive registry event %s", registryEvent.String())
				// 如果是服务掉线，发出提醒记录到数据库
				switch registryEvent.Type {
				case registry.SERVICE_REMOVE:
					store.LogServiceDown(ctx, registryEvent.ServiceInfo)
				}
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
