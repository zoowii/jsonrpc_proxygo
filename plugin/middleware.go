package plugin

import "github.com/zoowii/jsonrpc_proxygo/rpc"

type OnConnectionCont interface {
	OnConnection(session *rpc.ConnectionSession) error
}

type OnConnectionClosedCont interface {
	OnConnectionClosed(session *rpc.ConnectionSession) error
}

type Middleware interface {
	Name() string

	OnStart() error

	NextMiddleware() Middleware
	SetNextMiddleware(next Middleware)

	// return (continue bool, err error)
	OnConnection(session *rpc.ConnectionSession) error
	OnConnectionClosed(session *rpc.ConnectionSession) error

	OnWebSocketFrame(session *rpc.JSONRpcRequestSession, messageType int, message []byte) error
	OnRpcRequest(session *rpc.JSONRpcRequestSession) error
	OnRpcResponse(session *rpc.JSONRpcRequestSession) error

	ProcessRpcRequest(session *rpc.JSONRpcRequestSession) error
}
