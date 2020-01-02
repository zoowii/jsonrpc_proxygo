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

const (
	RPC_INTERNAL_ERROR = 10001

	RPC_UPSTREAM_CONNECTION_CLOSED_ERROR = 50001
	RPC_UPSTREAM_TIMEOUT_ERROR = 50002
)

func NewJSONRpcResponseError(code int, message string, data interface{}) *JSONRpcResponseError {
	return &JSONRpcResponseError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

type JSONRpcResponse struct {
	Id uint64 `json:"id"`
	JSONRpc string `json:"jsonrpc,omitempty"`
	Error *JSONRpcResponseError `json:"error,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func NewJSONRpcResponse(id uint64, result interface{}, err *JSONRpcResponseError) *JSONRpcResponse {
	return &JSONRpcResponse{
		Id:      id,
		JSONRpc: "2.0",
		Error:   err,
		Result:  result,
	}
}

func DecodeJSONRPCRequest(message []byte) (req *JSONRpcRequest, err error) {
	req = new(JSONRpcRequest)
	err = json.Unmarshal(message, &req)
	if err != nil {
		return
	}
	return
}

func EncodeJSONRPCResponse(res *JSONRpcResponse) (data []byte, err error) {
	data, err = json.Marshal(res)
	return
}

func DecodeJSONRPCResponse(message []byte) (req *JSONRpcResponse, err error) {
	req = new(JSONRpcResponse)
	err = json.Unmarshal(message, &req)
	if err != nil {
		return
	}
	return
}
