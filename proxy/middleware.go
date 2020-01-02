package proxy

type Middleware interface {
	Name() string
	// return (continue bool, err error)
	OnConnection(session *ConnectionSession) (bool, error)
	OnConnectionClosed(session *ConnectionSession) (bool, error)

	OnWebSocketFrame(session *JSONRpcRequestSession, messageType int, message []byte) (bool, error)
	OnJSONRpcRequest(session *JSONRpcRequestSession) (bool, error)
	OnJSONRpcResponse(session *JSONRpcRequestSession) (bool, error)

	ProcessJSONRpcRequest(session *JSONRpcRequestSession) (bool, error)
}

type MiddlewareChain struct {
	Middlewares []Middleware
}

func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{Middlewares:nil}
}

func (chain *MiddlewareChain) Append(middleware Middleware) *MiddlewareChain {
	chain.Middlewares = append(chain.Middlewares, middleware)
	return chain
}

func (chain *MiddlewareChain) InsertHead(middlewares ...Middleware) *MiddlewareChain {
	for _, middleware := range middlewares {
		chain.Append(middleware)
		count := len(chain.Middlewares)
		items := make([]Middleware, count)
		for i, j := 0, count-1; i <= j; i, j = i+1, j-1 {
			items[i], items[j] = chain.Middlewares[j], chain.Middlewares[i]
		}
		chain.Middlewares = items
	}
	return chain
}

func (chain *MiddlewareChain) OnConnection(session *ConnectionSession) (next bool, err error) {
	for _, m := range chain.Middlewares {
		next, err = m.OnConnection(session)
		if err != nil {
			return
		}
		if !next {
			break
		}
	}
	return
}

func (chain *MiddlewareChain) OnConnectionClosed(session *ConnectionSession) (next bool, err error) {
	for _, m := range chain.Middlewares {
		if m == nil {
			panic("null middleware")
			continue
		}
		next, err = m.OnConnectionClosed(session)
		if err != nil {
			return
		}
		if !next {
			break
		}
	}
	return
}

func (chain *MiddlewareChain) OnWebSocketFrame(session *JSONRpcRequestSession,
	messageType int, message []byte) (next bool, err error) {
	for _, m := range chain.Middlewares {
		next, err = m.OnWebSocketFrame(session, messageType, message)
		if err != nil {
			return
		}
		if !next {
			break
		}
	}
	return
}

func (chain *MiddlewareChain) OnJSONRpcRequest(session *JSONRpcRequestSession) (next bool, err error) {
	for _, m := range chain.Middlewares {
		next, err = m.OnJSONRpcRequest(session)
		if err != nil {
			return
		}
		if !next {
			break
		}
	}
	return
}

func (chain *MiddlewareChain) OnJSONRpcResponse(session *JSONRpcRequestSession) (next bool, err error) {
	for _, m := range chain.Middlewares {
		next, err = m.OnJSONRpcResponse(session)
		if err != nil {
			return
		}
		if !next {
			break
		}
	}
	return
}

func (chain *MiddlewareChain) ProcessJSONRpcRequest(session *JSONRpcRequestSession) (next bool, err error) {
	for _, m := range chain.Middlewares {
		next, err = m.ProcessJSONRpcRequest(session)
		if err != nil {
			return
		}
		if !next {
			break
		}
	}
	return
}