package providers

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketJsonRpcProvider struct {
	endpoint string
	websocketPath string
	middlewareChain *plugin.MiddlewareChain
}

func NewWebSocketJsonRpcProvider(endpoint string, websocketPath string) *WebSocketJsonRpcProvider {
	return &WebSocketJsonRpcProvider{
		endpoint:      endpoint,
		websocketPath: websocketPath,
		middlewareChain: nil,
	}
}

func (provider *WebSocketJsonRpcProvider) SetMiddlewareChain(chain *plugin.MiddlewareChain) {
	provider.middlewareChain = chain
}

func (provider *WebSocketJsonRpcProvider) serverHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	connSession := rpc.NewConnectionSession(w, r, c)
	defer connSession.Close()
	defer provider.middlewareChain.OnConnectionClosed(connSession)
	// must ensure middleware chain not change after calling OnConnection,
	// otherwise some removed middlewares may not call OnConnectionClosed
	if connErr := provider.middlewareChain.OnConnection(connSession); connErr != nil {
		log.Warn("OnConnection error", connErr)
		return
	}
	ctx := context.Background()
	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			case <- connSession.ConnectionDone:
				return
			case rpcDispatch := <- connSession.RpcRequestsDispatchChannel:
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
			case pack := <- connSession.RequestConnectionWriteChan:
				if pack == nil {
					return
				}
				err := c.WriteMessage(pack.MessageType, pack.Message)
				if err != nil {
					log.Println("write websocket frame error", err)
					return
				}
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if !utils.IsClosedOrGoingAwayCloseError(err) {
				log.Warn("read from source connection error:", err)
			}
			break
		}

		rpcSession := rpc.NewJSONRpcRequestSession(connSession)

		err = provider.middlewareChain.OnWebSocketFrame(rpcSession, mt, message)
		if err != nil {
			log.Warn("OnWebSocketFrame error", err)
			continue
		}
		switch mt {
		case websocket.CloseMessage:
			_ = c.Close()
			return
		}
		log.Debugf("recv: %s\n", message)
		if mt == websocket.BinaryMessage {
			// binary message should be processed by middlewares, not treated as jsonrpc request
			continue
		}
		rpcReq, err := rpc.DecodeJSONRPCRequest(message)
		if err != nil {
			log.Warn("jsonrpc request error", err)
			continue
		}
		rpcSession.Request = rpcReq
		rpcSession.RequestBytes = message
		err = provider.middlewareChain.OnJSONRpcRequest(rpcSession)
		if err != nil {
			log.Warn("OnRpcRequest error", err)
			continue
		}
		go func() {
			err = provider.middlewareChain.ProcessJSONRpcRequest(rpcSession)
			if err != nil {
				log.Warn("ProcessRpcRequest error", err)
				return
			}
			rpcRes := rpcSession.Response
			if rpcRes == nil {
				log.Error("empty jsonrpc response, maybe no valid middleware added")
				return
			}
			err = provider.middlewareChain.OnJSONRpcResponse(rpcSession)
			if err != nil {
				log.Warn("OnRpcResponse error", err)
				return
			}
			resBytes, err := rpc.EncodeJSONRPCResponse(rpcRes)
			if err != nil {
				log.Error("encodeJSONRPCResponse err", err)
				return
			}
			connSession.RequestConnectionWriteChan <- rpc.NewWebSocketPack(websocket.TextMessage, resBytes)
		}()
	}
}

func (provider *WebSocketJsonRpcProvider) ListenAndServe() (err error) {
	if provider.middlewareChain == nil {
		err = errors.New("please set provider.middlewareChain before ListenAndServe")
		return
	}
	wrappedHandler := func (w http.ResponseWriter, r *http.Request) {
		provider.serverHandler(w, r)
	}
	http.HandleFunc(provider.websocketPath, wrappedHandler)
	return http.ListenAndServe(provider.endpoint, nil)
}