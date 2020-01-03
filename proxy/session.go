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

type WebSocketPack struct {
	MessageType int
	Message []byte
}

func NewWebSocketPack(messageType int, message []byte) *WebSocketPack {
	return &WebSocketPack{
		MessageType: messageType,
		Message:     message,
	}
}

// TODO: 每一层可以做异步处理而不是同步处理（比如load balance)的middleware都有一个input chan和output chan，收到请求时
// input chan接收到数据以及output chan接收到一个待返回的slot(可以不同时),接收到input chan时去处理，处理后等output chan有空slot时填入返回数据

type ConnectionSession struct {
	RequestConnection *websocket.Conn
	HttpResponse http.ResponseWriter
	HttpRequest *http.Request
	ConnectionDone chan struct{}
	RequestConnectionWriteChan chan *WebSocketPack

	// rpc fields shared in connection session
	RpcRequestsMap map[uint64]chan *JSONRpcResponse // rpc request id => channel notify of *JSONRpcResponse
	RpcRequestsToDispatch chan *JSONRpcRequestSession // rpc requests to dispatch to processors(eg. upstream) to process
	RpcResponsesFromDispatchResult chan *JSONRpcRequestSession // rpc responses result from dispatched processsors

	// same base middleware shared fields

	// upstream middleware shared fields in connection session
	SelectedUpstreamTarget       *string
	UpstreamTargetConnection     *websocket.Conn
	UpstreamTargetConnectionDone chan struct{}
	UpstreamRpcRequestsChan      chan *JSONRpcRequestBundle

}

func (connSession *ConnectionSession) Close() {
	close(connSession.ConnectionDone)
	connSession.ConnectionDone = nil
	close(connSession.RequestConnectionWriteChan)
	connSession.RequestConnectionWriteChan = nil
	close(connSession.RpcRequestsToDispatch)
	connSession.RpcRequestsToDispatch = nil
	close(connSession.RpcResponsesFromDispatchResult)
	connSession.RpcResponsesFromDispatchResult = nil
}

func NewConnectionSession(w http.ResponseWriter, r *http.Request, requestConn *websocket.Conn) *ConnectionSession {
	return &ConnectionSession{
		RequestConnection: requestConn,
		RequestConnectionWriteChan: make(chan *WebSocketPack, 1000),
		HttpResponse: w,
		HttpRequest: r,
		ConnectionDone: make(chan struct{}),
		RpcRequestsMap: make(map[uint64] chan *JSONRpcResponse),
		RpcRequestsToDispatch: make(chan *JSONRpcRequestSession, 1000),
		RpcResponsesFromDispatchResult: make(chan *JSONRpcRequestSession, 1000),
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