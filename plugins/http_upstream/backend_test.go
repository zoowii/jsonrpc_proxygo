package http_upstream

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"testing"
)

func mockRpcConnection() *rpc.ConnectionSession {
	sess := rpc.NewConnectionSession()
	return sess
}

func mockRpcRequest(sess *rpc.ConnectionSession, method string, params []interface{}) *rpc.JSONRpcRequestSession {
	reqSess := rpc.NewJSONRpcRequestSession(sess)
	req := rpc.JSONRpcRequest{
		Id:      1,
		JSONRpc: "2.0",
		Method:  method,
		Params:  params,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.Fatalln(err)
	}
	reqSess.FillRpcRequest(&req, reqBytes)
	return reqSess
}

func TestNewHttpUpstreamMiddleware(t *testing.T) {
	utils.SetLogLevel("DEBUG")

	m, err := NewHttpUpstreamMiddleware(HttpDefaultTargetEndpoint("http://127.0.0.1:3000/api"))
	if err != nil {
		log.Fatalln(err)
	}
	log.Info("http backend default target endpoint is " + m.options.defaultTargetEndpoint)
	assert.True(t, err == nil)
	assert.Equal(t, m.Name(), "http-upstream")

	err = m.OnStart()
	assert.True(t, err == nil)

	sess := mockRpcConnection()

	err = m.OnConnection(sess)
	assert.True(t, err == nil)
	defer func() {
		err = m.OnConnectionClosed(sess)
		assert.True(t, err == nil)
	}()

	reqSess := mockRpcRequest(sess, "hello", []interface{}{"world"})
	err = m.OnRpcRequest(reqSess)
	assert.True(t, err == nil)

	err = m.ProcessRpcRequest(reqSess)
	assert.True(t, err == nil)

	err = m.OnRpcResponse(reqSess)
	assert.True(t, err == nil)

	// check rpc response
	rpcResp := reqSess.Response
	assert.True(t, rpcResp != nil)
	if rpcResp.Error != nil {
		log.Info("rpc error ", rpcResp.Error)
	}
	assert.True(t, rpcResp.Error == nil)
	log.Info("rpc response " + rpcResp.Result.(string))
	assert.True(t, rpcResp.Result.(string) == "Hello, world, this is response from server")
}
