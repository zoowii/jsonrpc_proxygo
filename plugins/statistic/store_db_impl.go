package statistic

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sony/sonyflake"
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
	dbUrl string
	db    *sql.DB
	sf    *sonyflake.Sonyflake
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
	traceId := reqSession.Request.Id // TODO
	rpcRequestId := fmt.Sprintf("%d", reqSession.Request.Id)
	rpcMethodName := reqSession.Request.Method
	var rpcRequestParams string
	if includeDebug {
		rpcRequestParams = utils.JsonDumpsToStringSilently(reqSession.Request.Params, "")
	}
	rpcResponseError := ""
	rpcResponseResult := ""
	targetServer := "" // TODO
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
	traceId := reqSession.Request.Id // TODO
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
	targetServer := "" // TODO
	_, err = stmt.Exec(id, annotation, traceId, rpcRequestId, rpcMethodName, rpcRequestParams,
		rpcResponseError, rpcResponseResult, targetServer)
	if err != nil {
		return
	}
}

const metricDbStoreName = "db"

func (store *metricDbStore) Name() string {
	return metricDbStoreName
}

func (store *metricDbStore) Init() error {
	db, err := createConn(store.dbUrl)
	if err != nil {
		return err
	}
	store.db = db
	return nil
}

func newMetricDbStore(dbUrl string) *metricDbStore {
	sonyFlakeSettings := sonyflake.Settings{
		StartTime: time.Now(),
	}
	sf := sonyflake.NewSonyflake(sonyFlakeSettings)
	return &metricDbStore{
		dbUrl: dbUrl,
		sf:    sf,
	}
}
