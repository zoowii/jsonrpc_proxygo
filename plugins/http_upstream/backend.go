package http_upstream

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	pluginsCommon "github.com/zoowii/jsonrpc_proxygo/plugins/common"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"io/ioutil"
	"net/http"
	"time"
)

var log = utils.GetLogger("http_upstream")

type HttpUpstreamMiddleware struct {
	plugin.MiddlewareAdapter
	DefaultTargetEndpoint string

	options *httpUpstreamMiddlewareOptions
}

func NewHttpUpstreamMiddleware(argOptions ...common.Option) (*HttpUpstreamMiddleware, error) {
	mOptions := &httpUpstreamMiddlewareOptions{
		upstreamTimeout:       30 * time.Second,
		defaultTargetEndpoint: "",
	}
	for _, o := range argOptions {
		o(mOptions)
	}
	m := &HttpUpstreamMiddleware{
		MiddlewareAdapter: plugin.MiddlewareAdapter{},
		options:           mOptions,
	}
	return m, nil
}

func (m *HttpUpstreamMiddleware) Name() string {
	return "http-upstream"
}

func (m *HttpUpstreamMiddleware) OnStart() (err error) {
	log.Info("http upstream plugin starting")
	return m.NextOnStart()
}

func (m *HttpUpstreamMiddleware) OnConnection(session *rpc.ConnectionSession) (err error) {
	log.Debugln("http upstream plugin on new connection")
	return m.NextOnConnection(session)
}

func (m *HttpUpstreamMiddleware) OnConnectionClosed(session *rpc.ConnectionSession) (err error) {
	// call next first
	err = m.NextOnConnectionClosed(session)
	log.Debugln("http upstream plugin on connection closed")
	return
}

func (m *HttpUpstreamMiddleware) getTargetEndpoint(session *rpc.ConnectionSession) (target string, err error) {
	return pluginsCommon.GetSelectedUpstreamTargetEndpoint(session, &m.DefaultTargetEndpoint)
}

func (m *HttpUpstreamMiddleware) OnRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	defer func() {
		if err == nil {
			err = m.NextOnJSONRpcRequest(session)
		}
	}()
	targetEndpoint, err := m.getTargetEndpoint(session.Conn)
	if err != nil {
		return
	}
	log.Debugln("http stream receive rpc request for backend " + targetEndpoint)
	// create response future before to use in ProcessRpcRequest
	session.RpcResponseFutureChan = make(chan *rpc.JSONRpcResponse, 1)
	rpcRequest := session.Request
	rpcRequestBytes, err := json.Marshal(rpcRequest)
	if err != nil {
		// TODO: write error rpc response
		return
	}
	log.Debugln("rpc request " + string(rpcRequestBytes))
	go func() {
		resp, err := http.Post(targetEndpoint, "application/json", bytes.NewReader(rpcRequestBytes))
		if err != nil {
			// TODO: write error rpc response
			return
		}
		defer resp.Body.Close()
		respMsg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// TODO: write error rpc response
			return
		}
		log.Debugln("backend rpc response " + string(respMsg))
		rpcRes, err := rpc.DecodeJSONRPCResponse(respMsg)
		if err != nil {
			return
		}
		if rpcRes == nil {
			err = errors.New("invalid jsonrpc response format from http upstream: " + string(respMsg))
			return
		}
		session.RpcResponseFutureChan <- rpcRes
	}()

	return
}

func (m *HttpUpstreamMiddleware) OnRpcResponse(session *rpc.JSONRpcRequestSession) (err error) {
	defer func() {
		if err == nil {
			err = m.NextOnJSONRpcResponse(session)
		}
	}()
	// TODO
	return
}

func (m *HttpUpstreamMiddleware) ProcessRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	defer func() {
		if err == nil {
			err = m.NextProcessJSONRpcRequest(session)
		}
	}()
	if session.Response != nil {
		return
	}
	rpcRequest := session.Request
	rpcRequestId := rpcRequest.Id
	requestChan := session.RpcResponseFutureChan
	if requestChan == nil {
		err = errors.New("can't find rpc request channel to process")
		return
	}

	var rpcRes *rpc.JSONRpcResponse
	select {
	case <-time.After(m.options.upstreamTimeout):
		rpcRes = rpc.NewJSONRpcResponse(rpcRequestId, nil,
			rpc.NewJSONRpcResponseError(rpc.RPC_UPSTREAM_CONNECTION_CLOSED_ERROR,
				"upstream target connection closed", nil))
	case <-session.Conn.UpstreamTargetConnectionDone:
		rpcRes = rpc.NewJSONRpcResponse(rpcRequestId, nil,
			rpc.NewJSONRpcResponseError(rpc.RPC_UPSTREAM_CONNECTION_CLOSED_ERROR,
				"upstream target connection closed", nil))
	case rpcRes = <-requestChan:
		// do nothing, just receive rpcRes
	}
	session.Response = rpcRes
	return
}

func (m *HttpUpstreamMiddleware) OnWebSocketFrame(session *rpc.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	defer func() {
		if err == nil {
			err = m.NextOnWebSocketFrame(session, messageType, message)
		}
	}()
	return
}
