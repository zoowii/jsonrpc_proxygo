package proxy

import (
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
	defer server.MiddlewareChain.OnConnectionClosed(w, r)

	if _, connErr := server.MiddlewareChain.OnConnection(w, r); connErr != nil {
		log.Println("OnConnection error", connErr)
		return
	}
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		switch mt {
		case websocket.PingMessage:
			c.WriteMessage(websocket.PongMessage, []byte("pong"))
		case websocket.CloseMessage:
			c.Close()
			return
		}
		rpcSession := NewJSONRpcRequestSession()
		rpcSession.HttpRequest = r
		rpcSession.HttpResponse = w
		_, err = server.MiddlewareChain.OnWebSocketFrame(w, r, mt, message)
		if err != nil {
			log.Println("OnWebSocketFrame error", err)
			continue
		}
		log.Printf("recv: %s", message)
		if mt == websocket.BinaryMessage {
			// binary message should be processed by middlewares, not treated as jsonrpc request
			continue
		}
		rpcReq, err := decodeJSONRPCRequest(message)
		if err != nil {
			log.Println("jsonrpc request error", err)
			continue
		}
		rpcSession.Request = rpcReq
		_, err = server.MiddlewareChain.OnJSONRpcRequest(rpcSession)
		if err != nil {
			log.Println("OnJSONRpcRequest error", err)
			continue
		}
		 _, err = server.MiddlewareChain.ProcessJSONRpcRequest(rpcSession)
		if err != nil {
			log.Println("ProcessJSONRpcRequest error", err)
			continue
		}
		rpcRes := rpcSession.Response
		if rpcRes == nil {
			log.Println("empty jsonrpc response, maybe no valid middleware added")
			continue
		}
		resBytes, err := encodeJSONRPCResponse(rpcRes)
		if err != nil {
			log.Println("encodeJSONRPCResponse err", err)
			continue
		}
		err = c.WriteMessage(mt, resBytes)
		if err != nil {
			log.Println("write message error", err)
			break
		}
	}
}

/**
 * Start the proxy server http service
 */
func (server *ProxyServer) Start() {
	log.Println("starting proxy server at " + server.Addr)
	wrappedHandler := func (w http.ResponseWriter, r *http.Request) {
		server.serverHandler(w, r)
	}
	http.HandleFunc(server.WebSocketPath, wrappedHandler)
	log.Fatal(http.ListenAndServe(server.Addr, nil))
} 