package statistic

/**
 * statistic middleware
 * async calculate statistic metrics and publish to admin users
 * TODO: publish metrics. now just dump metrics to log interval
 */

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"sync"
	"time"
)

var log = utils.GetLogger("statistic")

type StatisticMiddleware struct {
	proxy.MiddlewareAdapter
	rpcRequestsReceived chan *proxy.JSONRpcRequestSession
	rpcResponsesReceived chan *proxy.JSONRpcRequestSession

	globalRpcMethodsCount *utils.MemoryCache

	hourlyLock sync.RWMutex
	hourlyStartTime time.Time
	hourlyRpcMethodsCount *utils.MemoryCache
}

func NewStatisticMiddleware() *StatisticMiddleware {
	const maxRpcChannelSize = 10000

	return &StatisticMiddleware{
		rpcRequestsReceived: make(chan *proxy.JSONRpcRequestSession, maxRpcChannelSize),
		rpcResponsesReceived: make(chan *proxy.JSONRpcRequestSession, maxRpcChannelSize),
		globalRpcMethodsCount: utils.NewMemoryCache(),
		hourlyStartTime: time.Now(),
		hourlyRpcMethodsCount: utils.NewMemoryCache(),
	}
}

func (middleware *StatisticMiddleware) Name() string {
	return "statistic"
}

func getMethodNameForRpcStatistic(session *proxy.JSONRpcRequestSession) string {
	if session.MethodNameForCache != nil {
		return *session.MethodNameForCache // cache name is more acurrate for statistic
	}
	return session.Request.Method
}

func (middleware *StatisticMiddleware) OnStart() (err error) {
	go func() {
		ctx := context.Background()

		dumpIntervalOpened := false
		dumpTick := time.Tick(60 * time.Second)
		for {
			select {
			case <- ctx.Done():
				return
			case <- dumpTick:
				if !dumpIntervalOpened {
					continue
				}
				// notify user every some time
				log.Info("start dump statistic info")
				globalStatJson, err := middleware.globalRpcMethodsCount.DumpItems()
				if err != nil {
					log.Error("dump globalRpcMethodsCount error", err)
					continue
				}
				hourlyStatJson, err := middleware.hourlyRpcMethodsCount.DumpItems()
				if err != nil {
					log.Error("dump hourlyRpcMethodsCount error", err)
					continue
				}
				log.Infof("globalRpcMethodsCount: %s", string(globalStatJson))
				log.Infof("hourlyRpcMethodsCount: %s", string(hourlyStatJson))
			case reqSession := <- middleware.rpcRequestsReceived:
				methodNameForStatistic := getMethodNameForRpcStatistic(reqSession)

				// update global rpc methods called count
				_, ok := middleware.globalRpcMethodsCount.Get(methodNameForStatistic)
				if ok {
					_ = middleware.globalRpcMethodsCount.Increment(methodNameForStatistic, 1)
				} else {
					middleware.globalRpcMethodsCount.SetDefault(methodNameForStatistic, 1)
				}

				// update hourly rpc methods called count
				func() {
					middleware.hourlyLock.Lock()
					defer middleware.hourlyLock.Unlock()
					now := time.Now()
					if now.Sub(middleware.hourlyStartTime) > 1 * time.Hour {
						middleware.hourlyStartTime = now
						middleware.hourlyRpcMethodsCount.Flush() // delete all items
					}
					_, ok := middleware.hourlyRpcMethodsCount.Get(methodNameForStatistic)
					if ok {
						_ = middleware.hourlyRpcMethodsCount.Increment(methodNameForStatistic, 1)
					} else {
						middleware.hourlyRpcMethodsCount.SetDefault(methodNameForStatistic, 1)
					}
				}()
			case resSession := <- middleware.rpcResponsesReceived:
				_ = resSession // TODO
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
	return middleware.NextOnStart()
}

func (middleware *StatisticMiddleware) OnConnection(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnection(session)
}

func (middleware *StatisticMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *StatisticMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}
func (middleware *StatisticMiddleware) OnRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	middleware.rpcRequestsReceived <- session
	return middleware.NextOnJSONRpcRequest(session)
}
func (middleware *StatisticMiddleware) OnRpcResponse(session *proxy.JSONRpcRequestSession) (err error) {
	err = middleware.NextOnJSONRpcResponse(session)
	middleware.rpcResponsesReceived <- session
	return
}

func (middleware *StatisticMiddleware) ProcessRpcRequest(session *proxy.JSONRpcRequestSession) (err error) {
	return middleware.NextProcessJSONRpcRequest(session)
}
