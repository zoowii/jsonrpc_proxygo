package dashboard

import (
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"net/http"
)

var log = utils.GetLogger("dashboard-plugin")

type DashboardMiddleware struct {
	plugin.MiddlewareAdapter
	mOptions *dashboardOptions
}

func NewDashboardMiddleware(options ...common.Option) *DashboardMiddleware {
	mOptions := newDashBoardOptions()
	for _, o := range options {
		if o == nil {
			continue
		}
		o(mOptions)
	}
	return &DashboardMiddleware{
		mOptions: mOptions,
	}
}

func (m *DashboardMiddleware) Name() string {
	return "dashboard"
}

func (m *DashboardMiddleware) createDashboardWebHandler() http.Handler {
	r := m.mOptions.Registry
	store := m.mOptions.Store
	createDashboardApis(r, store)
	return nil
}

func (middleware *DashboardMiddleware) OnStart() (err error) {
	mOptions := middleware.mOptions
	endpoint := mOptions.Endpoint
	if len(endpoint) < 1 {
		return middleware.NextOnStart()
	}
	log.Infof("dashboard plugin listening endpoint %s", endpoint)
	go func() {
		handler := middleware.createDashboardWebHandler()
		err := http.ListenAndServe(endpoint, handler)
		if err != nil {
			log.Fatalf("dashboard server error %s", err.Error())
		}
	}()
	return middleware.NextOnStart()
}

func (middleware *DashboardMiddleware) OnConnection(session *rpc.ConnectionSession) (err error) {
	return middleware.NextOnConnection(session)
}

func (middleware *DashboardMiddleware) OnConnectionClosed(session *rpc.ConnectionSession) (err error) {
	return middleware.NextOnConnectionClosed(session)
}

func (middleware *DashboardMiddleware) OnWebSocketFrame(session *rpc.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	return middleware.NextOnWebSocketFrame(session, messageType, message)
}
func (middleware *DashboardMiddleware) OnRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcRequest(session)
}
func (middleware *DashboardMiddleware) OnRpcResponse(session *rpc.JSONRpcRequestSession) (err error) {
	return middleware.NextOnJSONRpcResponse(session)
}

func (middleware *DashboardMiddleware) ProcessRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	return middleware.NextProcessJSONRpcRequest(session)
}
