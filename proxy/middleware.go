package proxy

type OnConnectionCont interface {
	OnConnection(session *ConnectionSession) error
}

type OnConnectionClosedCont interface {
	OnConnectionClosed(session *ConnectionSession) error
}

type Middleware interface {
	Name() string

	OnStart() error

	NextMiddleware() Middleware
	SetNextMiddleware(next Middleware)

	// return (continue bool, err error)
	OnConnection(session *ConnectionSession) error
	OnConnectionClosed(session *ConnectionSession) error

	OnWebSocketFrame(session *JSONRpcRequestSession, messageType int, message []byte) error
	OnRpcRequest(session *JSONRpcRequestSession) error
	OnRpcResponse(session *JSONRpcRequestSession) error

	ProcessRpcRequest(session *JSONRpcRequestSession) error
}
