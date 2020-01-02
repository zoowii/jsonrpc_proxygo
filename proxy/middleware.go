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
	addCount := len(middlewares)
	if addCount>0 {
		count := len(chain.Middlewares)
		items := make([]Middleware, count + addCount)
		for i:=0;i<addCount;i++ {
			items[i]=middlewares[i]
		}
		for i:=0;i<count;i++ {
			items[i+addCount] = chain.Middlewares[i]
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

// TODO: OnJSONRpcResponse, OnConnectionClosed的调用顺序应该反向，因为没用middleware的单纯的wrap的方式
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