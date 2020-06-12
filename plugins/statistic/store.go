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
	Id                uint64     `json:"id"`
	Annotation        string     `json:"annotation"`
	TraceId           string     `json:"traceId"`
	RpcRequestId      string     `json:"rpcRequestId"`
	RpcMethodName     string     `json:"rpcMethodName"`
	RpcRequestParams  string     `json:"rpcRequestParams"`
	RpcResponseError  string     `json:"rpcResponseError"`
	RpcResponseResult string     `json:"rpcResponseResult"`
	TargetServer      string     `json:"targetServer"`
	LogTime           *time.Time `json:"logTime"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type RequestSpanListVo struct {
	Items []*RequestSpanVo `json:"items"`
	Total uint             `json:"total"`
}

type ServiceLogVo struct {
	Id          uint64     `json:"id"`
	ServiceName string     `json:"serviceName"`
	Url         string     `json:"url"`
	DownTime    *time.Time `json:"downTime"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type ServiceLogListVo struct {
	Items []*ServiceLogVo `json:"items"`
	Total uint            `json:"total"`
}

type ServiceHealthVo struct {
	Id          uint64        `json:"id"`
	ServiceName string        `json:"serviceName"`
	ServiceUrl  string        `json:"serviceUrl"`
	ServiceHost string        `json:"serviceHost"`
	Rtt         int64 `json:"rtt"` // milliseconds
	Connected   bool          `json:"connected"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
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
	QueryServiceDownLogs(ctx context.Context, offset int, limit int) (*ServiceLogListVo, error)

	UpdateServiceHostPing(ctx context.Context, service *registry.Service, rtt time.Duration, connected bool)
	QueryServiceHealthByUrl(ctx context.Context, service *registry.Service) (*ServiceHealthVo, error)

	DumpStatInfo() (dump *StatData, err error)
	addRpcMethodCall(methodName string)
}
