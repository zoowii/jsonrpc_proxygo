package ws_upstream

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/zoowii/jsonrpc_proxygo/plugins/common"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"time"
)

var log = utils.GetLogger("upstream")

type WsUpstreamMiddleware struct {
	UpstreamTimeout time.Duration
	DefaultTargetEndpoint string
}

func NewWsUpstreamMiddleware(defaultTargetEndpoint string) *WsUpstreamMiddleware {
	return &WsUpstreamMiddleware{
		UpstreamTimeout: 30 * time.Second,
		DefaultTargetEndpoint: defaultTargetEndpoint,
	}
}

func (middleware *WsUpstreamMiddleware) Name() string {
	return "ws-upstream"
}

// watch UpstreamTargetConnection's data
func (middleware *WsUpstreamMiddleware) watchUpstreamConnectionResponseAndToDispatchRpcRequests(session *proxy.ConnectionSession) {
	go func() {
		for {
			select {
			case <- session.UpstreamTargetConnectionDone:
				return
			case <- session.ConnectionDone:
				return
			case rpcRequestSession := <- session.RpcRequestsToDispatch:
				if rpcRequestSession == nil {
					return
				}
				rpcRequest := rpcRequestSession.Request
				rpcRequestId := rpcRequest.Id
				newChan := rpcRequestSession.RpcResponseFutureChan
				if old, ok := session.RpcRequestsMap[rpcRequestId]; ok {
					close(old)
				}
				session.RpcRequestsMap[rpcRequestId] = newChan
			case rpcRequestSession := <- session.RpcResponsesFromDispatchResult:
				if rpcRequestSession == nil {
					break
				}
				rpcRequest := rpcRequestSession.Request
				rpcRequestId := rpcRequest.Id
				rpcRequestSession.RpcResponseFutureChan = nil
				if resChan, ok := session.RpcRequestsMap[rpcRequestId]; ok {
					close(resChan)
					delete(session.RpcRequestsMap, rpcRequestId)
				}
			case req := <- session.UpstreamRpcRequestsChan:
				if req == nil {
					return
				}
				messageType := req.MessageType
				rpcRequest := req.Request
				rpcRequestBytes := req.Data
				targetConn := session.UpstreamTargetConnection
				if targetConn == nil {
					return
				}
				switch messageType {
				case websocket.PingMessage:
					_ = targetConn.WriteMessage(messageType, rpcRequestBytes)
					return
				case websocket.PongMessage:
					_ = targetConn.WriteMessage(messageType, rpcRequestBytes)
					return
				case websocket.BinaryMessage:
					_ = targetConn.WriteMessage(messageType, rpcRequestBytes)
					return
				case websocket.CloseMessage:
					_ = targetConn.WriteMessage(messageType, rpcRequestBytes)
					return
				}
				if rpcRequest == nil {
					log.Printf("null upstream channel request of message type %d\n", messageType)
					continue
				}
				err := targetConn.WriteMessage(websocket.TextMessage, rpcRequestBytes)
				if err != nil {
					log.Println("upstream write message error", err)
					// TODO: notify server.go to close the origin connection
					//close(session.ConnectionDone)
					return
				}
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

func (middleware *WsUpstreamMiddleware) OnStart() (err error) {
	return
}

func (middleware *WsUpstreamMiddleware) OnConnection(session *proxy.ConnectionSession) (next bool, err error) {
	if session.UpstreamTargetConnection != nil {
		err = errors.New("when OnConnection, session has connected to upstream target before")
		return
	}
	targetEndpoint, err := common.GetSelectedUpstreamTargetEndpoint(session, &middleware.DefaultTargetEndpoint)
	if err != nil {
		return
	}

	session.UpstreamRpcRequestsChan = make(chan *proxy.JSONRpcRequestBundle, 1000)

	go func() {
		log.Debugf("connecting to %s\n", targetEndpoint)
		c, _, err := websocket.DefaultDialer.Dial(targetEndpoint, nil)
		if err != nil {
			log.Println("dial:", err)
			return
		}
		log.Debugf("connected to %s\n", targetEndpoint)
		session.UpstreamTargetConnection = c
		session.UpstreamTargetConnectionDone = make(chan struct{})
		defer close(session.UpstreamTargetConnectionDone)

		middleware.watchUpstreamConnectionResponseAndToDispatchRpcRequests(session)

		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				_ = middleware.OnTargetWebsocketError(session, err)
				return
			}
			targetNext, err := middleware.OnTargetWebSocketFrame(session, messageType, message)
			if err != nil {
				log.Println("upstream target OnTargetWebSocketFrame error: ", err)
				continue
			}
			if !targetNext {
				break
			}
		}
	}()

	next = true
	return
}

func (middleware *WsUpstreamMiddleware) OnConnectionClosed(session *proxy.ConnectionSession) (next bool, err error) {
	next = true
	if session.UpstreamTargetConnection != nil {
		err = session.UpstreamTargetConnection.Close()
		if err == nil {
			session.UpstreamTargetConnection = nil
		}
	}
	close(session.UpstreamRpcRequestsChan)
	return
}

func (middleware *WsUpstreamMiddleware) OnTargetWebSocketFrame(session *proxy.ConnectionSession,
	messageType int, message []byte) (next bool, err error) {
	next = true
	requestConnWriteChan := session.RequestConnectionWriteChan
	var rpcRes *proxy.JSONRpcResponse
	switch messageType {
	case websocket.CloseMessage:
		next = false
		requestConnWriteChan <- proxy.NewWebSocketPack(messageType, message)
	case websocket.PingMessage:
		requestConnWriteChan <- proxy.NewWebSocketPack(messageType, message)
	case websocket.PongMessage:
		requestConnWriteChan <- proxy.NewWebSocketPack(messageType, message)
	case websocket.TextMessage:
		// process target rpc response
		rpcRes, err = proxy.DecodeJSONRPCResponse(message)
		if err != nil {
			return
		}
		if rpcRes == nil {
			err = errors.New("invalid jsonrpc response format from upstream: " + string(message))
			return
		}
		rpcRequestId := rpcRes.Id
		if rpcReqChan, ok := session.RpcRequestsMap[rpcRequestId]; ok {
			rpcReqChan <- rpcRes
		}
	}
	return
}

func (middleware *WsUpstreamMiddleware) OnTargetWebsocketError(session *proxy.ConnectionSession, err error) error {
	if utils.IsClosedOrGoingAwayCloseError(err) {
		return nil
	}
	if err == nil {
		return nil
	}
	log.Println("upstream target websocket error:", err)
	return nil
}

func (middleware *WsUpstreamMiddleware) sendRequestToTargetConn(session *proxy.ConnectionSession, messageType int,
	message []byte, rpcRequest *proxy.JSONRpcRequest, rpcResponseFutureChan chan *proxy.JSONRpcResponse) {
	session.UpstreamRpcRequestsChan <- proxy.NewJSONRpcRequestBundle(messageType, message, rpcRequest, rpcResponseFutureChan)
}

func (middleware *WsUpstreamMiddleware) OnWebSocketFrame(session *proxy.JSONRpcRequestSession,
	messageType int, message []byte) (bool, error) {
	switch messageType {
	case websocket.PingMessage:
		middleware.sendRequestToTargetConn(session.Conn, messageType, message, nil, nil)
	case websocket.PongMessage:
		middleware.sendRequestToTargetConn(session.Conn, messageType, message, nil, nil)
	case websocket.BinaryMessage:
		middleware.sendRequestToTargetConn(session.Conn, messageType, message, nil, nil)
	case websocket.CloseMessage:
		middleware.sendRequestToTargetConn(session.Conn, messageType, message, nil, nil)
	}
	return true, nil
}
func (middleware *WsUpstreamMiddleware) OnJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	connSession := session.Conn
	rpcRequest := session.Request
	rpcRequestBytes := session.RequestBytes

	// create response future before to use in ProcessJSONRpcRequest
	session.RpcResponseFutureChan = make(chan *proxy.JSONRpcResponse)
	connSession.RpcRequestsToDispatch <- session

	middleware.sendRequestToTargetConn(connSession, websocket.TextMessage, rpcRequestBytes, rpcRequest, session.RpcResponseFutureChan)
	return
}
func (middleware *WsUpstreamMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	connSession := session.Conn
	defer func() {
		connSession.RpcResponsesFromDispatchResult <- session // notify connSession this rpc response is end
	}()
	responseBytes, jsonErr := json.Marshal(session.Response)
	if jsonErr == nil {
		log.Debugf("upstream response: %s", string(responseBytes))
	}
	return
}

func (middleware *WsUpstreamMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
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

	var rpcRes *proxy.JSONRpcResponse
	select {
	case <- time.After(middleware.UpstreamTimeout):
		rpcRes = proxy.NewJSONRpcResponse(rpcRequestId, nil,
			proxy.NewJSONRpcResponseError(proxy.RPC_UPSTREAM_CONNECTION_CLOSED_ERROR,
				"upstream target connection closed", nil))
	case <- session.Conn.UpstreamTargetConnectionDone:
		rpcRes = proxy.NewJSONRpcResponse(rpcRequestId, nil,
			proxy.NewJSONRpcResponseError(proxy.RPC_UPSTREAM_CONNECTION_CLOSED_ERROR,
				"upstream target connection closed", nil))
	case rpcRes = <- requestChan:
		// do nothing, just receive rpcRes
	}
	session.Response = rpcRes
	return
}