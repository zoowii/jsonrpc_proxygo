package statistic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sony/sonyflake"
	"github.com/zoowii/jsonrpc_proxygo/registry"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"time"
)

// createConn by dbUrl format like 'user:pass@tcp(ip:port)/dbName?param1=value1&param2=value2'
func createConn(dbUrl string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dbUrl)
	return db, err
}

func nextId(sf *sonyflake.Sonyflake) uint64 {
	id, err := sf.NextID()
	if err != nil {
		panic(err)
	}
	return id
}

type metricDbStore struct {
	BaseMetricStore
	dbUrl string
	db    *sql.DB
	sf    *sonyflake.Sonyflake
}

func newMetricDbStore(dbUrl string) *metricDbStore {
	sonyFlakeSettings := sonyflake.Settings{
		StartTime: time.Unix(0, 0),
	}
	sf := sonyflake.NewSonyflake(sonyFlakeSettings)
	return &metricDbStore{
		dbUrl: dbUrl,
		sf:    sf,
	}
}

func (store *metricDbStore) LogRequest(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool) {
	db := store.db
	if db == nil {
		return
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer func() {
		if err != nil {
			log.Errorf("tx error %s", err.Error())
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	stmt, err := tx.Prepare("insert into request_span (`id`, `annotation`, `trace_id`, `rpc_request_id`, " +
		"`rpc_method_name`, `rpc_request_params`, `rpc_response_error`, `rpc_response_result`, " +
		"`target_server`) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	id := nextId(store.sf)
	annotation := "sr"
	traceId := reqSession.Request.Id
	rpcRequestId := fmt.Sprintf("%d", reqSession.Request.Id)
	rpcMethodName := reqSession.Request.Method
	var rpcRequestParams string
	if includeDebug {
		rpcRequestParams = utils.JsonDumpsToStringSilently(reqSession.Request.Params, "")
	}
	rpcResponseError := ""
	rpcResponseResult := ""
	targetServer := reqSession.TargetServer
	_, err = stmt.Exec(id, annotation, traceId, rpcRequestId, rpcMethodName, rpcRequestParams,
		rpcResponseError, rpcResponseResult, targetServer)
	if err != nil {
		return
	}
}

func (store *metricDbStore) logResponse(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool) {
	db := store.db
	if db == nil {
		return
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer func() {
		if err != nil {
			log.Errorf("tx error %s", err.Error())
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	stmt, err := tx.Prepare("insert into request_span (`id`, `annotation`, `trace_id`, `rpc_request_id`, " +
		"`rpc_method_name`, `rpc_request_params`, `rpc_response_error`, `rpc_response_result`, " +
		"`target_server`) values (?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	id := nextId(store.sf)
	annotation := "ss"
	traceId := reqSession.Request.Id
	rpcRequestId := fmt.Sprintf("%d", reqSession.Request.Id)
	rpcMethodName := reqSession.Request.Method
	var rpcRequestParams string
	if includeDebug {
		rpcRequestParams = utils.JsonDumpsToStringSilently(reqSession.Request.Params, "")
	}
	response := reqSession.Response
	var rpcResponseError string
	if response != nil {
		if response.Error != nil {
			rpcResponseError = utils.JsonDumpsToStringSilently(response.Error, response.Error.Message)
		}
	} else {
		rpcResponseError = "no response"
	}
	var rpcResponseResult string
	if response != nil {
		if response.Error == nil {
			rpcResponseResult = utils.JsonDumpsToStringSilently(response.Result, "")
		}
	}
	targetServer := reqSession.TargetServer
	_, err = stmt.Exec(id, annotation, traceId, rpcRequestId, rpcMethodName, rpcRequestParams,
		rpcResponseError, rpcResponseResult, targetServer)
	if err != nil {
		return
	}
}

func (store *metricDbStore) QueryRequestSpanList(ctx context.Context, form *QueryLogForm) (result *RequestSpanListVo, err error) {
	db := store.db
	if db == nil {
		err = errors.New("metric db not init")
		return
	}
	totalRows, err := db.Query("select count(1) from request_span")
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer totalRows.Close()
	var total uint
	if totalRows.Next() {
		err = totalRows.Scan(&total)
		if err != nil {
			log.Warn("metric db error", err)
			return
		}
	}
	rows, err := db.Query("select `id`, `annotation`, `trace_id`, `rpc_request_id`, `rpc_method_name`,"+
		" `rpc_request_params`, `rpc_response_error`, `rpc_response_result`, `target_server`, `log_time`,"+
		" `create_at`, `update_at` from `request_span` order by `create_at` desc limit ?, ?", form.Offset, form.Limit)
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer rows.Close()
	list := &RequestSpanListVo{
		Items: make([]*RequestSpanVo, 0),
		Total: total,
	}
	for ; rows.Next(); {
		var item RequestSpanVo
		err = rows.Scan(&item.Id, &item.Annotation, &item.TraceId, &item.RpcRequestId, &item.RpcMethodName,
			&item.RpcRequestParams, &item.RpcResponseError, &item.RpcResponseResult, &item.TargetServer,
			&item.LogTime, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			log.Warn("metric db error", err)
			return
		}
		list.Items = append(list.Items, &item)
	}
	result = list
	return
}

func (store *metricDbStore) LogServiceDown(ctx context.Context, service *registry.Service) {
	db := store.db
	if db == nil {
		return
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer func() {
		if err != nil {
			log.Errorf("tx error %s", err.Error())
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	stmt, err := tx.Prepare("insert into service_log (`id`, `service_name`, `url`, `down_time`) values (?, ?, ?, ?)")
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	id := nextId(store.sf)
	log.Infof("new service_log id %d", id)
	serviceName := service.Name
	serviceUrl := service.Url
	downTime := time.Now().UTC()
	_, err = stmt.Exec(id, serviceName, serviceUrl, downTime)
	if err != nil {
		return
	}
}

func (store *metricDbStore) QueryServiceDownLogs(ctx context.Context, offset int, limit int) (result *ServiceLogListVo, err error) {
	db := store.db
	if db == nil {
		err = errors.New("metric db not init")
		return
	}
	totalRows, err := db.Query("select count(1) from service_log where `down_time` is not null")
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer totalRows.Close()
	var total uint
	if totalRows.Next() {
		err = totalRows.Scan(&total)
		if err != nil {
			log.Warn("metric db error", err)
			return
		}
	}
	rows, err := db.Query("select `id`, `service_name`, `url`, `down_time`,"+
		" `create_at`, `update_at` from `service_log` where `down_time` is not null order by `create_at` desc limit ?, ?", offset, limit)
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer rows.Close()
	list := &ServiceLogListVo{
		Items: make([]*ServiceLogVo, 0),
		Total: total,
	}
	for ; rows.Next(); {
		var item ServiceLogVo
		err = rows.Scan(&item.Id, &item.ServiceName, &item.Url, &item.DownTime, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			log.Warn("metric db error", err)
			return
		}
		list.Items = append(list.Items, &item)
	}
	result = list
	return
}

func (store *metricDbStore) UpdateServiceHostPing(ctx context.Context, service *registry.Service, rtt time.Duration, connected bool) {
	db := store.db
	if db == nil {
		return
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer func() {
		if err != nil {
			log.Errorf("tx error %s", err.Error())
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	stmt, err := tx.Prepare("insert INTO service_health (`id`, `service_name`, `service_url`, `service_host`, `rtt`, `connected`)" +
		" VALUES (?, ?, ?, ?, ?, ?)" +
		" on DUPLICATE  key update `service_url`=?, `service_host`=?, `rtt`=?, `connected`=?")
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	id := nextId(store.sf)
	log.Infof("replace service_health id %d of host %s", id, service.Host)
	serviceName := service.Name
	serviceUrl := service.Url
	host := service.Host
	connectedInt := 0
	if connected {
		connectedInt = 1
	}
	rttInt := rtt.Nanoseconds() / 1e6
	_, err = stmt.Exec(id, serviceName, serviceUrl, host, rttInt, connectedInt, serviceUrl, host, rttInt, connectedInt)
	if err != nil {
		return
	}
}

func (store *metricDbStore) QueryServiceHealthByUrl(ctx context.Context, service *registry.Service) (result *ServiceHealthVo, err error) {
	// 找到某个服务的ping状态
	db := store.db
	if db == nil {
		err = errors.New("metric db not init")
		return
	}
	serviceUrl := service.Url
	rows, err := db.Query("select `id`, `service_name`, `service_url`, `service_host`, `rtt`, `connected`,"+
		" `create_at`, `update_at` from `service_health` where service_url=?", serviceUrl)
	if err != nil {
		log.Warn("metric db error", err)
		return
	}
	defer rows.Close()
	if rows.Next() {
		var item ServiceHealthVo
		var connectedInt int
		err = rows.Scan(&item.Id, &item.ServiceName, &item.ServiceUrl, &item.ServiceHost, &item.Rtt, &connectedInt,
			&item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			log.Warn("metric db error", err)
			return
		}
		if connectedInt > 0 {
			item.Connected = true
		} else {
			item.Connected = false
		}
		result = &item
	}
	return
}

const metricDbStoreName = "db"

func (store *metricDbStore) Name() string {
	return metricDbStoreName
}

func (store *metricDbStore) Init() error {
	err := store.BaseMetricStore.Init()
	if err != nil {
		return err
	}
	db, err := createConn(store.dbUrl)
	if err != nil {
		return err
	}
	store.db = db
	return nil
}
