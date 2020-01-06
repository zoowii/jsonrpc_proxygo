package proxy

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

/**
 * ProxyServer: proxy server type
 */
type ProxyServer struct {
	Addr string
	WebSocketPath string // default "/"
	MiddlewareChain *MiddlewareChain
}

/**
 * NewProxyServer: init and return a new proxy server instance
 */
func NewProxyServer(addr string) *ProxyServer {
	server := &ProxyServer{
		Addr: addr,
		WebSocketPath: "/",
		MiddlewareChain: NewMiddlewareChain(),
	}
	return server
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var log = utils.GetLogger("server")

// TODO: websocket jsonrpc subscribe and unsubscribe

func (server *ProxyServer) serverHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	connSession := NewConnectionSession(w, r, c)
	defer connSession.Close()
	defer server.MiddlewareChain.OnConnectionClosed(connSession)
	// must ensure middleware chain not change after calling OnConnection,
	// otherwise some removed middlewares may not call OnConnectionClosed
	if connErr := server.MiddlewareChain.OnConnection(connSession); connErr != nil {
		log.Warn("OnConnection error", connErr)
		return
	}
	ctx := context.Background()
	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			case <- connSession.ConnectionDone:
				return
			case rpcDispatch := <- connSession.RpcRequestsDispatchChannel:
				if rpcDispatch == nil {
					return
				}
				rpcRequestSession := rpcDispatch.Data
				switch rpcDispatch.Type {
				case RPC_REQUEST_CHANGE_TYPE_ADD_REQUEST:
					rpcRequest := rpcRequestSession.Request
					rpcRequestId := rpcRequest.Id
					newChan := rpcRequestSession.RpcResponseFutureChan
					if old, ok := connSession.RpcRequestsMap[rpcRequestId]; ok {
						close(old)
					}
					connSession.RpcRequestsMap[rpcRequestId] = newChan
				case RPC_REQUEST_CHANGE_TYPE_ADD_RESPONSE:
					rpcRequest := rpcRequestSession.Request
					rpcRequestId := rpcRequest.Id
					rpcRequestSession.RpcResponseFutureChan = nil
					if resChan, ok := connSession.RpcRequestsMap[rpcRequestId]; ok {
						close(resChan)
						delete(connSession.RpcRequestsMap, rpcRequestId)
					}
				}
			case pack := <- connSession.RequestConnectionWriteChan:
				if pack == nil {
					return
				}
				err := c.WriteMessage(pack.MessageType, pack.Message)
				if err != nil {
					log.Println("write websocket frame error", err)
					return
				}
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if !utils.IsClosedOrGoingAwayCloseError(err) {
				log.Warn("read from source connection error:", err)
			}
			break
		}

		rpcSession := NewJSONRpcRequestSession(connSession)

		err = server.MiddlewareChain.OnWebSocketFrame(rpcSession, mt, message)
		if err != nil {
			log.Warn("OnWebSocketFrame error", err)
			continue
		}
		switch mt {
		case websocket.CloseMessage:
			_ = c.Close()
			return
		}
		log.Debugf("recv: %s\n", message)
		if mt == websocket.BinaryMessage {
			// binary message should be processed by middlewares, not treated as jsonrpc request
			continue
		}
		rpcReq, err := DecodeJSONRPCRequest(message)
		if err != nil {
			log.Warn("jsonrpc request error", err)
			continue
		}
		rpcSession.Request = rpcReq
		rpcSession.RequestBytes = message
		err = server.MiddlewareChain.OnJSONRpcRequest(rpcSession)
		if err != nil {
			log.Warn("OnRpcRequest error", err)
			continue
		}
		go func() {
			err = server.MiddlewareChain.ProcessJSONRpcRequest(rpcSession)
			if err != nil {
				log.Warn("ProcessRpcRequest error", err)
				return
			}
			rpcRes := rpcSession.Response
			if rpcRes == nil {
				log.Error("empty jsonrpc response, maybe no valid middleware added")
				return
			}
			err = server.MiddlewareChain.OnJSONRpcResponse(rpcSession)
			if err != nil {
				log.Warn("OnRpcResponse error", err)
				return
			}
			resBytes, err := EncodeJSONRPCResponse(rpcRes)
			if err != nil {
				log.Error("encodeJSONRPCResponse err", err)
				return
			}
			connSession.RequestConnectionWriteChan <- NewWebSocketPack(websocket.TextMessage, resBytes)
		}()
	}
}

func (server *ProxyServer) StartMiddlewares() error {
	return server.MiddlewareChain.OnStart()
}

/**
 * Start the proxy server http service
 */
func (server *ProxyServer) Start() {
	wrappedHandler := func (w http.ResponseWriter, r *http.Request) {
		server.serverHandler(w, r)
	}
	http.HandleFunc(server.WebSocketPath, wrappedHandler)
	log.Fatal(http.ListenAndServe(server.Addr, nil))
} 