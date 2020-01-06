package plugin

import "github.com/zoowii/jsonrpc_proxygo/rpc"

type MiddlewareAdapter struct {
	nextMiddleware Middleware
}

func (adapter *MiddlewareAdapter) NextMiddleware() Middleware {
	return adapter.nextMiddleware
}

func (adapter *MiddlewareAdapter) SetNextMiddleware(next Middleware) {
	adapter.nextMiddleware = next
}


func (middleware *MiddlewareAdapter) NextOnStart() (err error) {
	next := middleware.NextMiddleware()
	if next != nil {
		err = next.OnStart()
	}
	return
}

func (middleware *MiddlewareAdapter) NextOnConnection(session *rpc.ConnectionSession) (err error) {
	next := middleware.NextMiddleware()
	if next != nil {
		err = next.OnConnection(session)
	}
	return
}

func (middleware *MiddlewareAdapter) NextOnConnectionClosed(session *rpc.ConnectionSession) (err error) {
	next := middleware.NextMiddleware()
	if next != nil {
		err = next.OnConnectionClosed(session)
	}
	return
}

func (middleware *MiddlewareAdapter) NextOnWebSocketFrame(session *rpc.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	next := middleware.NextMiddleware()
	if next != nil {
		err = next.OnWebSocketFrame(session, messageType, message)
	}
	return
}
func (middleware *MiddlewareAdapter) NextOnJSONRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	next := middleware.NextMiddleware()
	if next != nil {
		err = next.OnRpcRequest(session)
	}
	return
}
func (middleware *MiddlewareAdapter) NextOnJSONRpcResponse(session *rpc.JSONRpcRequestSession) (err error) {
	next := middleware.NextMiddleware()
	if next != nil {
		err = next.OnRpcResponse(session)
	}
	return
}

func (middleware *MiddlewareAdapter) NextProcessJSONRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	next := middleware.NextMiddleware()
	if next != nil {
		err = next.ProcessRpcRequest(session)
	}
	return
}
