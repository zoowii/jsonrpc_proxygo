package proxy

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type JSONRpcRequestBundle struct {
	MessageType int
	Data []byte
	Request *JSONRpcRequest
	ResponseFutureChan chan *JSONRpcResponse
}

func NewJSONRpcRequestBundle(messageType int, data []byte, rpcRequest *JSONRpcRequest, rpcResponseFutureChan chan *JSONRpcResponse) *JSONRpcRequestBundle {
	return &JSONRpcRequestBundle{
		MessageType: messageType,
		Data:        data,
		Request:     rpcRequest,
		ResponseFutureChan: rpcResponseFutureChan,
	}
}

type ConnectionSession struct {
	RequestConnection *websocket.Conn
	HttpResponse http.ResponseWriter
	HttpRequest *http.Request
	ConnectionDone chan struct{}

	// rpc fields shared in connection session
	RpcRequestsMap map[uint64]chan *JSONRpcResponse // rpc request id => channel notify of *JSONRpcResponse

	// same base middleware shared fields

	// upstream middleware shared fields in connection session
	SelectedUpstreamTarget       *string
	UpstreamTargetConnection     *websocket.Conn
	UpstreamTargetConnectionDone chan struct{}
	UpstreamRpcRequestsChan      chan *JSONRpcRequestBundle

}

func (connSession *ConnectionSession) Close() {
	close(connSession.ConnectionDone)
}

func NewConnectionSession(w http.ResponseWriter, r *http.Request, requestConn *websocket.Conn) *ConnectionSession {
	return &ConnectionSession{
		RequestConnection: requestConn,
		HttpResponse: w,
		HttpRequest: r,
		ConnectionDone: make(chan struct{}),
		RpcRequestsMap: make(map[uint64] chan *JSONRpcResponse),
	}
}

type JSONRpcRequestSession struct {
	Conn *ConnectionSession
	RequestBytes []byte
	Request *JSONRpcRequest
	Response *JSONRpcResponse
	Parameters map[string]interface{}

	RpcResponseFutureChan chan *JSONRpcResponse

	// cache middleware shared fields
	ResponseSetByCache bool

	// before_cache middleware shared fields
	MethodNameForCache *string // only used to find cache in cache middleware
}

func NewJSONRpcRequestSession(conn *ConnectionSession) *JSONRpcRequestSession {
	return &JSONRpcRequestSession{
		Conn: conn,
		ResponseSetByCache: false,
	}
}