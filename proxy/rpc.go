package proxy

import "encoding/json"

type JSONRpcRequest struct {
	Id uint64 `json:"id"`
	JSONRpc string `json:"jsonrpc,omitempty"`
	Method string `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

type JSONRpcResponseError struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}

type JSONRpcResponse struct {
	Id uint64 `json:"id"`
	JSONRpc string `json:"jsonrpc,omitempty"`
	Error *JSONRpcResponseError `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func decodeJSONRPCRequest(message []byte) (req *JSONRpcRequest, err error) {
	req = new(JSONRpcRequest)
	err = json.Unmarshal(message, &req)
	if err != nil {
		return
	}
	return
}

func encodeJSONRPCResponse(res *JSONRpcResponse) (data []byte, err error) {
	data, err = json.Marshal(res)
	return
}