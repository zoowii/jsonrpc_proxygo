package main

import (
	"github.com/zoowii/jsonrpc_proxygo/plugins/ws_upstream"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"log"
)

func main() {
	addr := "localhost:5000"
	log.Println("to start proxy server on " + addr)
	server := proxy.NewProxyServer(addr)
	targetEndpoint := "ws://localhost:3000"
	server.MiddlewareChain.InsertHead(
		ws_upstream.NewWsUpstreamMiddleware(targetEndpoint),
	)
	server.Start()
}
