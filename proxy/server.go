package proxy

import (
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/providers"
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

/**
 * Start the proxy server http service
 */
func (server *ProxyServer) Start() {
	if server.Provider == nil {
		log.Fatalln("please set provider to ProxyServer before start")
		return
	}
	server.Provider.SetMiddlewareChain(server.MiddlewareChain)
	log.Fatal(server.Provider.ListenAndServe())
} 