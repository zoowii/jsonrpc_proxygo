package proxy

import (
	"github.com/gorilla/websocket"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/providers"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
)

var log = utils.GetLogger("server")

/**
 * ProxyServer: proxy server type
 */
type ProxyServer struct {
	MiddlewareChain *plugin.MiddlewareChain
	Provider providers.RpcProvider
}

/**
 * NewProxyServer: init and return a new proxy server instance
 */
func NewProxyServer(provider providers.RpcProvider) *ProxyServer {
	server := &ProxyServer{
		MiddlewareChain: plugin.NewMiddlewareChain(),
		Provider: provider,
	}
	return server
}

func (server *ProxyServer) StartMiddlewares() error {
	return server.MiddlewareChain.OnStart()
}

func (server *ProxyServer) NotifyNewConnection(connSession *rpc.ConnectionSession) error {
	return server.MiddlewareChain.OnConnection(connSession)
}

func (server *ProxyServer) OnConnectionClosed(connSession *rpc.ConnectionSession) error {
	// must ensure middleware chain not change after calling OnConnection,
	// otherwise some removed middlewares may not call OnConnectionClosed
	return server.MiddlewareChain.OnConnectionClosed(connSession)
}

func (server *ProxyServer) OnRawRequestMessage(connSession *rpc.ConnectionSession, rpcSession *rpc.JSONRpcRequestSession,
	messageType int, message []byte) error {
	return server.MiddlewareChain.OnWebSocketFrame(rpcSession, messageType, message)
}

func (server *ProxyServer) OnRpcRequest(connSession *rpc.ConnectionSession, rpcSession *rpc.JSONRpcRequestSession) (err error) {
	err = server.MiddlewareChain.OnJSONRpcRequest(rpcSession)
	if err != nil {
		log.Warn("OnRpcRequest error", err)
		return
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
		resBytes, err := rpc.EncodeJSONRPCResponse(rpcRes)
		if err != nil {
			log.Error("encodeJSONRPCResponse err", err)
			return
		}
		connSession.RequestConnectionWriteChan <- rpc.NewWebSocketPack(websocket.TextMessage, resBytes)
	}()
	return
}

/**
 * Start the proxy server http service
 */
func (server *ProxyServer) Start() {
	if server.Provider == nil {
		log.Fatalln("please set provider to ProxyServer before start")
		return
	}
	server.Provider.SetRpcProcessor(server)
	log.Fatal(server.Provider.ListenAndServe())
} 