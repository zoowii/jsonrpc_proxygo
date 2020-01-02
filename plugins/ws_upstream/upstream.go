package ws_upstream

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/zoowii/jsonrpc_proxygo/plugins/common"
	"github.com/zoowii/jsonrpc_proxygo/proxy"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"log"
	"time"
)

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
	return "upstream"
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
	utils.Debugf("connecting to %s\n", targetEndpoint)
	// TODO: connect target in a goroutine
	// TODO: when target is wss url
	c, _, err := websocket.DefaultDialer.Dial(targetEndpoint, nil)
	if err != nil {
		log.Println("dial:", err)
		return
	}
	utils.Debugf("connected to %s\n", targetEndpoint)
	session.UpstreamTargetConnection = c
	session.UpstreamTargetConnectionDone = make(chan struct{})
	session.UpstreamRpcRequestsChan = make(chan *proxy.JSONRpcRequestBundle, 1000)
	// watch UpstreamTargetConnection's data
	go func(middleware *WsUpstreamMiddleware, session *proxy.ConnectionSession) {
		defer close(session.UpstreamTargetConnectionDone)
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
	}(middleware, session)

	go func() {
		for {
			select {
			case <- session.UpstreamTargetConnectionDone:
				break
			case <- session.ConnectionDone:
				break
			case req := <- session.UpstreamRpcRequestsChan:
				if req == nil {
					return
				}
				messageType := req.MessageType
				rpcRequest := req.Request
				rpcRequestBytes := req.Data
				targetConn := session.UpstreamTargetConnection
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
				rpcRequestId := rpcRequest.Id

				newChan := req.ResponseFutureChan
				if old, ok := session.RpcRequestsMap[rpcRequestId]; ok {
					close(old)
				}
				session.RpcRequestsMap[rpcRequestId] = newChan
				err := targetConn.WriteMessage(websocket.TextMessage, rpcRequestBytes)
				if err != nil {
					log.Println("upstream write message error", err)
					// TODO: notify server.go to close the origin connection
					//close(session.ConnectionDone)
					break
				}
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
		close(session.UpstreamRpcRequestsChan)
	}
	return
}

func (middleware *WsUpstreamMiddleware) OnTargetWebSocketFrame(session *proxy.ConnectionSession,
	messageType int, message []byte) (next bool, err error) {
	next = true
	requestConn := session.RequestConnection
	var rpcRes *proxy.JSONRpcResponse
	switch messageType {
	case websocket.CloseMessage:
		next = false
		err = requestConn.WriteMessage(messageType, message) // TODO: send msg to channel to let main goroutine to process socket write
	case websocket.PingMessage:
		err = requestConn.WriteMessage(messageType, message)
	case websocket.PongMessage:
		err = requestConn.WriteMessage(messageType, message)
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

	middleware.sendRequestToTargetConn(connSession, websocket.TextMessage, rpcRequestBytes, rpcRequest, session.RpcResponseFutureChan)
	return
}
func (middleware *WsUpstreamMiddleware) OnJSONRpcResponse(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	return
}

func (middleware *WsUpstreamMiddleware) ProcessJSONRpcRequest(session *proxy.JSONRpcRequestSession) (next bool, err error) {
	next = true
	rpcRequest := session.Request
	rpcRequestId := rpcRequest.Id
	rpcRequestsMap := session.Conn.RpcRequestsMap
	requestChan := session.RpcResponseFutureChan
	if requestChan == nil {
		err = errors.New("can't find rpc request channel to process")
		return
	}
	defer func() {
		close(session.RpcResponseFutureChan)
		delete(rpcRequestsMap, rpcRequestId)
	}()
	var rpcRes *proxy.JSONRpcResponse
	select {
	case <- session.Conn.UpstreamTargetConnectionDone:
		rpcRes = proxy.NewJSONRpcResponse(rpcRequestId, nil,
			proxy.NewJSONRpcResponseError(proxy.RPC_UPSTREAM_CONNECTION_CLOSED_ERROR,
				"upstream target connection closed", nil))
	case rpcRes = <- requestChan:

	case <- time.After(middleware.UpstreamTimeout):
		rpcRes = proxy.NewJSONRpcResponse(rpcRequestId, nil,
			proxy.NewJSONRpcResponseError(proxy.RPC_UPSTREAM_CONNECTION_CLOSED_ERROR,
				"upstream target connection closed", nil))
	}
	session.Response = rpcRes
	return
}