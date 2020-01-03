package proxy

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"log"
	"net/http"

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
	if _, connErr := server.MiddlewareChain.OnConnection(connSession); connErr != nil {
		log.Println("OnConnection error", connErr)
		return
	}
	ctx := context.Background()
	go func() {
		for {
			select {
			case <- ctx.Done():
				break
			case <- connSession.ConnectionDone:
				break
			case pack := <- connSession.RequestConnectionWriteChan:
				if pack == nil {
					break
				}
				err := c.WriteMessage(pack.MessageType, pack.Message)
				if err != nil {
					log.Println("write websocket frame error", err)
					break
				}
			}
		}
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		rpcSession := NewJSONRpcRequestSession(connSession)

		_, err = server.MiddlewareChain.OnWebSocketFrame(rpcSession, mt, message)
		if err != nil {
			log.Println("OnWebSocketFrame error", err)
			continue
		}
		switch mt {
		case websocket.CloseMessage:
			_ = c.Close()
			return
		}
		utils.Debugf("[server]recv: %s\n", message)
		if mt == websocket.BinaryMessage {
			// binary message should be processed by middlewares, not treated as jsonrpc request
			continue
		}
		rpcReq, err := DecodeJSONRPCRequest(message)
		if err != nil {
			log.Println("jsonrpc request error", err)
			continue
		}
		rpcSession.Request = rpcReq
		rpcSession.RequestBytes = message
		_, err = server.MiddlewareChain.OnJSONRpcRequest(rpcSession)
		if err != nil {
			log.Println("OnJSONRpcRequest error", err)
			continue
		}
		go func() {
			_, err = server.MiddlewareChain.ProcessJSONRpcRequest(rpcSession)
			if err != nil {
				log.Println("ProcessJSONRpcRequest error", err)
				return
			}
			rpcRes := rpcSession.Response
			if rpcRes == nil {
				log.Println("empty jsonrpc response, maybe no valid middleware added")
				return
			}
			_, err = server.MiddlewareChain.OnJSONRpcResponse(rpcSession)
			if err != nil {
				log.Println("OnJSONRpcResponse error", err)
				return
			}
			resBytes, err := EncodeJSONRPCResponse(rpcRes)
			if err != nil {
				log.Println("encodeJSONRPCResponse err", err)
				return
			}
			connSession.RequestConnectionWriteChan <- NewWebSocketPack(websocket.TextMessage, resBytes)
		}()
	}
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