jsonrpc_proxygo
===================

http/websocket JSONRPC server proxy with middleware mechanism.

you can add middlewares like upstream, cache, load balance, log, tracing, heartbeat, statistic, rate-limit, etc.

# Supported Middlewares

* exposed http jsonrpc and websocket jsonrpc interfaces(providers)
* upstream: dispatch jsonrpc(based on websocket) to backend endpoints
* load-balance: use WeightedRound-Robin algorithm to select one endpoint to use in upstream middleware
* cache: cache some jsonrpc method's responses by jsonrpc method name and some params for some time
* before-cache: extract some jsonrpc params to cache key to use in cache middleware
* statistic: calculate statistic metrics of the jsonrpc services. It works async and won't block the service
* rate-limit

# Usage

* install by go get
```
go get github.com/zoowii/jsonrpc_proxygo

# and then cd to the jsonrpc_proxygo as work dir
go build
```

* copy `server.sample.json` to `server.json` and update the config json file

```
// example of server.json
{
  "resolver": "consul://127.0.0.1:8500",
  "config_file_resolver": "consul://127.0.0.1:8500/v1/kv/jsonrpc_proxy.config",
  "endpoint": "127.0.0.1:5000",
  "provider": "websocket",
  "log": {
    "level": "INFO",
    "output_file": "logs/jsonrpc_proxygo.log"
  },
  "plugins": {
    "upstream": {
      "upstream_endpoints": [
        {
          "url": "ws://127.0.0.1:3000", "weight": 1
        },
        {
          "url": "wss://127.0.0.1:4000", "weight": 2
        }
      ]
    },
    "caches": [
      { "name": "dummyMethod", "expire_seconds": 5 },
      { "name": "call", "paramsForCache": [2, "getSomeInfoMethod"],  "expire_seconds": 5 }
    ],
    "before_cache_configs": [
      {"method": "call", "fetch_cache_key_from_params_count": 2}
    ],
    "statistic": {
      "start": true
    },
    "disable": {
      "start": true,
      "disabled_rpc_methods": [
        "stop"
      ]
    },
    "rate_limit": {
      "start": true,
      "connection_rate": 10000,
      "rpc_rate": 1000000
    }
  }
}

```

* run the jsonrpc_proxygo proxy server

```
./jsonrpc_proxygo -config server.json

# sample output maybe looks like:
{"level":"info","module":"main","msg":"to start proxy server on 127.0.0.1:5000","time":""}
{"level":"info","module":"main","msg":"loaded middlewares are(count 5):\n","time":""}
{"level":"info","module":"main","msg":"\t- middleware statistic\n","time":""}
{"level":"info","module":"main","msg":"\t- middleware before_cache\n","time":""}
{"level":"info","module":"main","msg":"\t- middleware cache\n","time":""}
{"level":"info","module":"main","msg":"\t- middleware load_balance\n","time":""}
{"level":"info","module":"main","msg":"\t- middleware upstream\n","time":""}

```