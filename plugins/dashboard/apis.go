package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/plugins/statistic"
	"github.com/zoowii/jsonrpc_proxygo/registry"
	"io/ioutil"
	"net/http"
)

type apiHandlers struct {
	store statistic.MetricStore
	r registry.Registry
}

func newApiHandlers(store statistic.MetricStore, r registry.Registry) *apiHandlers {
	return &apiHandlers{
		store: store,
		r: r,
	}
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

func sendResult(writer http.ResponseWriter, result interface{}) {
	mBytes, err := json.Marshal(result)
	if err != nil {
		sendErrorResponse(writer, err)
		return
	}
	_, err = writer.Write(mBytes)
	if err != nil {
		log.Errorf("api send result error %s", err.Error())
	}
}

func readJsonBody(request *http.Request, value interface{}) (err error) {
	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(bodyBytes, value)
	return
}

func (h *apiHandlers) statisticApi(writer http.ResponseWriter, request *http.Request) {
	// 统计摘要数据
	log.Info("receive /api/statistic")
	store := h.store
	r := h.r

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

	sendResult(writer, statInfo)
}

func (h *apiHandlers) listRequestSpanApi(writer http.ResponseWriter, request *http.Request) {
	log.Info("receive list_request_span api")
	store := h.store
	if store == nil {
		sendErrorResponse(writer, errors.New("metric store not init"))
		return
	}
	form := &statistic.QueryLogForm{
		Offset: 0,
		Limit: 20,
	}
	err := readJsonBody(request, form)
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

	sendResult(writer, reqSpanList)
}

func (h *apiHandlers) listServiceDownLogsApi(writer http.ResponseWriter, request *http.Request) {
	log.Info("receive list_service_down_logs api")
	store := h.store
	if store == nil {
		sendErrorResponse(writer, errors.New("metric store not init"))
		return
	}
	type formType struct {
		Offset int `json:"offset"`
		Limit int `json:"limit"`
	}
	form := &formType{
		Offset: 0,
		Limit: 20,
	}
	err := readJsonBody(request, form)
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
	reqSpanList, err := store.QueryServiceDownLogs(context.Background(), form.Offset, form.Limit)
	if err != nil {
		sendErrorResponse(writer, err)
		return
	}

	sendResult(writer, reqSpanList)
}

func (h *apiHandlers) queryServiceHealthApi(writer http.ResponseWriter, request *http.Request) {
	log.Info("queryServiceHealthApi called")
	store := h.store
	if store == nil {
		sendErrorResponse(writer, errors.New("metric store not init"))
		return
	}
	type formType struct {
		Name string `json:"name"`
		Url string `json:"url"`
	}
	form := &formType{}
	err := readJsonBody(request, form)
	if err != nil {
		sendErrorResponse(writer, err)
		return
	}
	if len(form.Name) < 1 && len(form.Url) < 1 {
		sendErrorResponse(writer, errors.New("empty name and url form"))
		return
	}
	result, err := store.QueryServiceHealthByUrl(context.Background(), &registry.Service{
		Name: form.Name,
		Url: form.Url,
	})
	if err != nil {
		sendErrorResponse(writer, err)
		return
	}
	sendResult(writer, result)
}

func (h *apiHandlers) wrapApi(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func (writer http.ResponseWriter, request *http.Request) {
		allowCors(&writer, request)
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusOK)
			return
		}
		handlerFunc(writer, request)
	}
}

func createDashboardApis(r registry.Registry, store statistic.MetricStore) {
	hs := newApiHandlers(store, r)
	http.HandleFunc("/api/statistic", hs.wrapApi(hs.statisticApi))
	http.HandleFunc("/api/list_request_span", hs.wrapApi(hs.listRequestSpanApi))
	http.HandleFunc("/api/list_service_down_logs", hs.wrapApi(hs.listServiceDownLogsApi))
	http.HandleFunc("/api/query_service_health", hs.wrapApi(hs.queryServiceHealthApi))
}
