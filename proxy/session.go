package proxy

import "net/http"

type JSONRpcRequestSession struct {
	HttpResponse http.ResponseWriter
	HttpRequest *http.Request
	Request *JSONRpcRequest
	Response *JSONRpcResponse
	Parameters map[string]interface{}
}

func NewJSONRpcRequestSession() *JSONRpcRequestSession {
	return &JSONRpcRequestSession{}
}