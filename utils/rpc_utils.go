package utils

import (
	"encoding/json"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
)

func CloneJSONRpcResponse(source *proxy.JSONRpcResponse) (result *proxy.JSONRpcResponse, err error) {
	bytes, err := json.Marshal(source)
	if err != nil {
		return
	}
	result = new(proxy.JSONRpcResponse)
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return
	}
	return
}
