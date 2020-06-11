package providers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"io/ioutil"
	"net/http"
	"time"
)

type HttpJsonRpcProviderOptions struct {
	TimeoutSeconds uint32
}

type HttpJsonRpcProvider struct {
	endpoint     string
	path         string
	options      *HttpJsonRpcProviderOptions
	rpcProcessor RpcProviderProcessor
}

func NewHttpJsonRpcProvider(endpoint string, path string, options *HttpJsonRpcProviderOptions) *HttpJsonRpcProvider {
	if options == nil {
		log.Fatalln("null HttpJsonRpcProviderOptions provided")
		return nil
	}
	return &HttpJsonRpcProvider{
		endpoint:     endpoint,
		path:         path,
		rpcProcessor: nil,
		options:      options,
	}
}

func (provider *HttpJsonRpcProvider) SetRpcProcessor(processor RpcProviderProcessor) {
	provider.rpcProcessor = processor
}

func sendErrorResponse(w http.ResponseWriter, err error, errCode int, requestId uint64) {
	resErr := rpc.NewJSONRpcResponseError(errCode, err.Error(), nil)
	errRes := rpc.NewJSONRpcResponse(requestId, nil, resErr)
	errResBytes, encodeErr := json.Marshal(errRes)
	if encodeErr != nil {
		_, _ = w.Write(errResBytes)
	}
}

func (provider *HttpJsonRpcProvider) asyncWatchMessagesToConnection(ctx context.Context, connSession *rpc.ConnectionSession,
	w http.ResponseWriter, r *http.Request, done chan struct{}) {
	go func() {
		defer func() {
			done <- struct{}{}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-connSession.ConnectionDone:
				return
			case rpcDispatch := <-connSession.RpcRequestsDispatchChannel:
				if rpcDispatch == nil {
					return
				}
				rpcRequestSession := rpcDispatch.Data
				switch rpcDispatch.Type {
				case rpc.RPC_REQUEST_CHANGE_TYPE_ADD_REQUEST:
					rpcRequest := rpcRequestSession.Request
					rpcRequestId := rpcRequest.Id
					newChan := rpcRequestSession.RpcResponseFutureChan
					if old, ok := connSession.RpcRequestsMap[rpcRequestId]; ok {
						close(old)
					}
					connSession.RpcRequestsMap[rpcRequestId] = newChan
				case rpc.RPC_REQUEST_CHANGE_TYPE_ADD_RESPONSE:
					rpcRequest := rpcRequestSession.Request
					rpcRequestId := rpcRequest.Id
					rpcRequestSession.RpcResponseFutureChan = nil
					if resChan, ok := connSession.RpcRequestsMap[rpcRequestId]; ok {
						close(resChan)
						delete(connSession.RpcRequestsMap, rpcRequestId)
					}
				}
			case pack := <-connSession.RequestConnectionWriteChan:
				if pack == nil {
					return
				}
				_, err := w.Write(pack.Message)
				if err != nil {
					log.Warn("write response error", err)
					return
				}
				return
			case <-time.After(time.Duration(provider.options.TimeoutSeconds) * time.Second):
				sendErrorResponse(w, errors.New("timeout"), rpc.RPC_RESPONSE_TIMEOUT_ERROR, 0)
				return
			}
		}
	}()
}

func (provider *HttpJsonRpcProvider) watchConnectionMessages(ctx context.Context, connSession *rpc.ConnectionSession, w http.ResponseWriter, r *http.Request) (err error) {
	body := r.Body
	defer body.Close()
	message, err := ioutil.ReadAll(body)
	if err != nil {
		return
	}
	log.Debugf("recv: %s\n", message)
	rpcSession := rpc.NewJSONRpcRequestSession(connSession)

	messageType := 0

	err = provider.rpcProcessor.OnRawRequestMessage(connSession, rpcSession, messageType, message)
	if err != nil {
		log.Warn("OnRawRequestMessage error", err)
		return
	}
	rpcReq, err := rpc.DecodeJSONRPCRequest(message)
	if err != nil {
		err = errors.New("jsonrpc request error" + err.Error())
		return
	}
	rpcSession.FillRpcRequest(rpcReq, message)
	err = provider.rpcProcessor.OnRpcRequest(connSession, rpcSession)
	return
}

func (provider *HttpJsonRpcProvider) serverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		_, _ = w.Write([]byte("ok"))
		return
	}
	if r.Method != http.MethodPost {
		resErr := rpc.NewJSONRpcResponseError(rpc.RPC_INTERNAL_ERROR, "only support POST method", nil)
		errRes := rpc.NewJSONRpcResponse(0, nil, resErr)
		errResBytes, encodeErr := json.Marshal(errRes)
		if encodeErr != nil {
			_, _ = w.Write(errResBytes)
		}
		return
	}
	// TODO: 对于暴露http接口来说，backend upstream不对每次请求都启动一个websocket conn，而是共用一个连接池. 暂时每次都创建新连接
	connSession := rpc.NewConnectionSession()
	defer connSession.Close()
	defer provider.rpcProcessor.OnConnectionClosed(connSession)
	if connErr := provider.rpcProcessor.NotifyNewConnection(connSession); connErr != nil {
		log.Warn("OnConnection error", connErr)
		return
	}
	ctx := context.Background()

	err := provider.watchConnectionMessages(ctx, connSession, w, r)
	if err != nil {
		sendErrorResponse(w, err, rpc.RPC_INTERNAL_ERROR, 0)
		return
	}

	done := make(chan struct{})
	provider.asyncWatchMessagesToConnection(ctx, connSession, w, r, done)

	select {
	case <-done:
		break
	}
}

func (provider *HttpJsonRpcProvider) ListenAndServe() (err error) {
	if provider.rpcProcessor == nil {
		err = errors.New("please set provider.rpcProcessor before ListenAndServe")
		return
	}
	wrappedHandler := func(w http.ResponseWriter, r *http.Request) {
		provider.serverHandler(w, r)
	}
	http.HandleFunc(provider.path, wrappedHandler)
	return http.ListenAndServe(provider.endpoint, nil)
}
