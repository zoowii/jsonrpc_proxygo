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
	allowCors(&writer, request)
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
	allowCors(&writer, request)
	store := h.store

	if request.Method == "OPTIONS" {
		writer.WriteHeader(http.StatusOK)
		return
	}
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


func createDashboardApis(r registry.Registry) {
	store := statistic.UsedMetricStore

	hs := newApiHandlers(store, r)
	http.HandleFunc("/api/statistic", hs.statisticApi)
	http.HandleFunc("/api/list_request_span", hs.listRequestSpanApi)

	// TODO: 更多的API
}
