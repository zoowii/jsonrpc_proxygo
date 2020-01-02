package main

import (
	"flag"
	"github.com/zoowii/jsonrpc_proxygo/config"
	"github.com/zoowii/jsonrpc_proxygo/plugins/ws_upstream"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"io/ioutil"
	"log"
)

func main() {
	// TODO: load config from yml config file
	configPath := flag.String("config", "server.json", "configuration file path(default server.json)")
	flag.Parse()
	configFileBytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalln(err)
		return
	}
	var configInfo config.ServerConfig
	err = config.UnmarshalServerConfigFromJson(configFileBytes, &configInfo)
	if err != nil {
		log.Fatalln(err)
		return
	}
	addr := configInfo.Endpoint
	log.Println("to start proxy server on " + addr)
	server := proxy.NewProxyServer(addr)
	// TODO: when config info not full
	targetEndpoint := configInfo.Plugins.Upstream.TargetEndpoints[0]
	server.MiddlewareChain.InsertHead(
		ws_upstream.NewWsUpstreamMiddleware(targetEndpoint),
	)
	server.Start()
}
