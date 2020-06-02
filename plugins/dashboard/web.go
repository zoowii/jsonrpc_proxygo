package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/plugins/statistic"
	"github.com/zoowii/jsonrpc_proxygo/registry"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"io/ioutil"
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

func (middleware *DashboardMiddleware) Name() string {
	return "dashboard"
}

func allowCors(writer *http.ResponseWriter, request *http.Request) {
	(*writer).Header().Add("Access-Control-Allow-Credentials", "true")
	(*writer).Header().Set("Access-Control-Allow-Origin", "*")
	(*writer).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*writer).Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-Auth-Token, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Access-Control-Allow-Methods, Content-Length")
}

func sendErrorResponse(writer http.ResponseWriter, e error) {
	writer.WriteHeader(http.StatusInternalServerError)
	m := struct{
		Error struct{
			Message string `json:"message"`
		} `json:"error"`
	}{}
	m.Error.Message = e.Error()
	mBytes, jsonErr := json.Marshal(m)
	if jsonErr != nil {
		log.Fatalln("json marshal error", jsonErr)
		return
	}
	_, writerErr := writer.Write(mBytes)
	if writerErr != nil {
		log.Fatalln("http write response error", writerErr)
		return
	}
	return
}

func (m *DashboardMiddleware) createDashboardWebHandler() http.Handler {
	store := statistic.UsedMetricStore
	r := m.mOptions.Registry
	http.HandleFunc("/api/statistic", func(writer http.ResponseWriter, request *http.Request) {
		// 统计摘要数据
		log.Info("receive /api/statistic")
		allowCors(&writer, request)

		if store == nil {
			sendErrorResponse(writer, errors.New("metric store not init"))
			return
		}
		statInfo, err := store.DumpStatInfo()
		if err != nil {
			sendErrorResponse(writer, err)
			return
		}

		if r != nil {
			services, err := r.ListServices()
			if err != nil {
				sendErrorResponse(writer, err)
				return
			}
			upstreamServices := make([]*registry.Service, 0)
			for _, s := range services {
				if s.Name == "upstream" {
					upstreamServices = append(upstreamServices, s)
				}
			}
			statInfo.UpstreamServices = upstreamServices
			statInfo.Services = services
		}

		mBytes, err := json.Marshal(statInfo)
		if err != nil {
			sendErrorResponse(writer, err)
			return
		}
		writer.Write(mBytes)
	})
	http.HandleFunc("/api/list_request_span", func(writer http.ResponseWriter, request *http.Request) {
		log.Info("receive list_request_span api")
		allowCors(&writer, request)
		if request.Method == "OPTIONS" {
			writer.WriteHeader(http.StatusOK)
			return
		}
		if store == nil {
			sendErrorResponse(writer, errors.New("metric store not init"))
			return
		}
		bodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			sendErrorResponse(writer, err)
			return
		}
		form := &statistic.QueryLogForm{
			Offset: 0,
			Limit: 20,
		}
		err = json.Unmarshal(bodyBytes, form)
		if err != nil {
			sendErrorResponse(writer, err)
			return
		}
		if form.Offset < 0 {
			form.Offset = 0
		}
		if form.Limit <= 0 {
			form.Limit = 20
		}
		reqSpanList, err := store.QueryRequestSpanList(context.Background(), form)
		if err != nil {
			sendErrorResponse(writer, err)
			return
		}

		mBytes, err := json.Marshal(reqSpanList)
		if err != nil {
			sendErrorResponse(writer, err)
			return
		}
		writer.Write(mBytes)

	})
	// TODO: 更多的API
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
