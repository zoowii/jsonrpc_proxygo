package providers

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
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
	rpcProcessor RpcProviderProcessor
}

func NewWebSocketJsonRpcProvider(endpoint string, websocketPath string) *WebSocketJsonRpcProvider {
	return &WebSocketJsonRpcProvider{
		endpoint:      endpoint,
		websocketPath: websocketPath,
		rpcProcessor: nil,
	}
}

func (provider *WebSocketJsonRpcProvider) SetRpcProcessor(processor RpcProviderProcessor) {
	provider.rpcProcessor = processor
}

func (provider *WebSocketJsonRpcProvider) asyncWatchMessagesToConnection(ctx context.Context, connSession *rpc.ConnectionSession, c *websocket.Conn) {
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
}

func (provider *WebSocketJsonRpcProvider) watchConnectionMessages(ctx context.Context, connSession *rpc.ConnectionSession, c *websocket.Conn) {
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			if !utils.IsClosedOrGoingAwayCloseError(err) {
				log.Warn("read from source connection error:", err)
			}
			break
		}

		rpcSession := rpc.NewJSONRpcRequestSession(connSession)

		err = provider.rpcProcessor.OnRawRequestMessage(connSession, rpcSession, mt, message)
		if err != nil {
			log.Warn("OnWebSocketFrame error", err)
			continue
		}
		switch mt {
		case websocket.CloseMessage:
			return
		}
		log.Debugf("recv: %s\n", message)
		if mt == websocket.BinaryMessage {
			// binary message should be processed by middlewares only, not treated as jsonrpc request
			continue
		}
		rpcReq, err := rpc.DecodeJSONRPCRequest(message)
		if err != nil {
			log.Warn("jsonrpc request error", err)
			continue
		}
		rpcSession.Request = rpcReq
		rpcSession.RequestBytes = message
		err = provider.rpcProcessor.OnRpcRequest(connSession, rpcSession)
		if err != nil {
			log.Warn(err)
			continue
		}
	}
}

func (provider *WebSocketJsonRpcProvider) serverHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn("upgrade:", err)
		return
	}
	defer c.Close()
	connSession := rpc.NewConnectionSession()
	defer connSession.Close()
	defer provider.rpcProcessor.OnConnectionClosed(connSession)
	if connErr := provider.rpcProcessor.NotifyNewConnection(connSession); connErr != nil {
		log.Warn("OnConnection error", connErr)
		return
	}
	ctx := context.Background()

	provider.asyncWatchMessagesToConnection(ctx, connSession, c)
	provider.watchConnectionMessages(ctx, connSession, c)
}

func (provider *WebSocketJsonRpcProvider) ListenAndServe() (err error) {
	if provider.rpcProcessor == nil {
		err = errors.New("please set provider.rpcProcessor before ListenAndServe")
		return
	}
	wrappedHandler := func (w http.ResponseWriter, r *http.Request) {
		provider.serverHandler(w, r)
	}
	http.HandleFunc(provider.websocketPath, wrappedHandler)
	return http.ListenAndServe(provider.endpoint, nil)
}
