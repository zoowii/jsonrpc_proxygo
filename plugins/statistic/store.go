package statistic

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/registry"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"time"
)

type QueryLogForm struct {
	Offset uint `json:"offset"`
	Limit  uint `json:"limit"`
}

type RequestSpanVo struct {
	Id                uint64    `json:"id"`
	Annotation        string    `json:"annotation"`
	TraceId           string    `json:"traceId"`
	RpcRequestId      string    `json:"rpcRequestId"`
	RpcMethodName     string    `json:"rpcMethodName"`
	RpcRequestParams  string    `json:"rpcRequestParams"`
	RpcResponseError  string    `json:"rpcResponseError"`
	RpcResponseResult string    `json:"rpcResponseResult"`
	TargetServer      string    `json:"targetServer"`
	LogTime           time.Time `json:"logTime"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type RequestSpanListVo struct {
	Items []*RequestSpanVo `json:"items"`
	Total uint             `json:"total"`
}

type MetricStore interface {
	Name() string
	Init() error
	// LogRequest store request info. if {includeDebug}==true, store request's content and debug info
	LogRequest(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool)
	// LogRequest store response info. if {includeDebug}==true, store response's content and debug info
	logResponse(ctx context.Context, reqSession *rpc.JSONRpcRequestSession, includeDebug bool)
	QueryRequestSpanList(ctx context.Context, form *QueryLogForm) (*RequestSpanListVo, error)

	LogServiceDown(ctx context.Context, service *registry.Service)

	DumpStatInfo() (dump *StatData, err error)
	addRpcMethodCall(methodName string)
}
