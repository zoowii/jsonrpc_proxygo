package ws_upstream

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	pluginsCommon "github.com/zoowii/jsonrpc_proxygo/plugins/common"
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"time"
)

var log = utils.GetLogger("upstream")

type WsUpstreamMiddleware struct {
	plugin.MiddlewareAdapter

	options *wsUpstreamMiddlewareOptions
}

func NewWsUpstreamMiddleware(argOptions ...common.Option) *WsUpstreamMiddleware {
	mOptions := &wsUpstreamMiddlewareOptions{
		upstreamTimeout:       30 * time.Second,
		defaultTargetEndpoint: "",
	}
	for _, o := range argOptions {
		o(mOptions)
	}
	return &WsUpstreamMiddleware{
		options: mOptions,
	}
}

func (middleware *WsUpstreamMiddleware) Name() string {
	return "ws-upstream"
}

// watch UpstreamTargetConnection's data
func (middleware *WsUpstreamMiddleware) watchUpstreamConnectionResponseAndToDispatchRpcRequests(session *rpc.ConnectionSession) {
	go func() {
		for {
			select {
			case <-session.UpstreamTargetConnectionDone:
				return
			case <-session.ConnectionDone:
				return
			case req := <-session.UpstreamRpcRequestsChan:
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
					log.Error("upstream write message error", err)
					// notify server.go to close the origin connection
					session.ConnectionDone <- struct{}{}
					return
				}
			default:
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

func (middleware *WsUpstreamMiddleware) OnStart() (err error) {
	log.Info("websocket upstream plugin starting")
	return middleware.NextOnStart()
}

func (middleware *WsUpstreamMiddleware) getTargetEndpoint(session *rpc.ConnectionSession) (target string, err error) {
	return pluginsCommon.GetSelectedUpstreamTargetEndpoint(session, &middleware.options.defaultTargetEndpoint)
}

func connectTargetEndpoint(targetEndpoint string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(targetEndpoint, nil)
	return conn, err
}

func (middleware *WsUpstreamMiddleware) OnConnection(session *rpc.ConnectionSession) (err error) {
	// TODO: upstream连接可以设置每个连接新建一个到upstream的连接，也可以选择从upstream connection pool中选择一个满足target的连接复用
	if session.UpstreamTargetConnection != nil {
		err = errors.New("when OnConnection, session has connected to upstream target before")
		return
	}
	targetEndpoint, err := middleware.getTargetEndpoint(session)
	if err != nil {
		return
	}

	session.UpstreamRpcRequestsChan = make(chan *rpc.JSONRpcRequestBundle, 1000)
	if session.SelectedUpstreamTarget == nil {
		session.SelectedUpstreamTarget = &targetEndpoint
	}

	go func() {
		log.Debugf("connecting to %s\n", targetEndpoint)
		c, err := connectTargetEndpoint(targetEndpoint)
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

	return middleware.NextOnConnection(session)
}

func (middleware *WsUpstreamMiddleware) OnConnectionClosed(session *rpc.ConnectionSession) (err error) {
	// call next first
	err = middleware.NextOnConnectionClosed(session)

	if session.UpstreamTargetConnection != nil {
		err = session.UpstreamTargetConnection.Close()
		if err == nil {
			session.UpstreamTargetConnection = nil
		}
	}
	close(session.UpstreamRpcRequestsChan)
	return
}

func (middleware *WsUpstreamMiddleware) OnTargetWebSocketFrame(session *rpc.ConnectionSession,
	messageType int, message []byte) (next bool, err error) {
	next = true
	requestConnWriteChan := session.RequestConnectionWriteChan
	var rpcRes *rpc.JSONRpcResponse
	switch messageType {
	case websocket.CloseMessage:
		next = false
		requestConnWriteChan <- rpc.NewMessagePack(messageType, message)
	case websocket.PingMessage:
		requestConnWriteChan <- rpc.NewMessagePack(messageType, message)
	case websocket.PongMessage:
		requestConnWriteChan <- rpc.NewMessagePack(messageType, message)
	case websocket.TextMessage:
		// process target rpc response
		rpcRes, err = rpc.DecodeJSONRPCResponse(message)
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

func (middleware *WsUpstreamMiddleware) OnTargetWebsocketError(session *rpc.ConnectionSession, err error) error {
	if utils.IsClosedOrGoingAwayCloseError(err) {
		return nil
	}
	if err == nil {
		return nil
	}
	log.Println("upstream target websocket error:", err)
	return nil
}

func (middleware *WsUpstreamMiddleware) sendRequestToTargetConn(session *rpc.ConnectionSession, messageType int,
	message []byte, rpcRequest *rpc.JSONRpcRequest, rpcResponseFutureChan chan *rpc.JSONRpcResponse) {
	session.UpstreamRpcRequestsChan <- rpc.NewJSONRpcRequestBundle(messageType, message, rpcRequest, rpcResponseFutureChan)
}

func (middleware *WsUpstreamMiddleware) OnWebSocketFrame(session *rpc.JSONRpcRequestSession,
	messageType int, message []byte) (err error) {
	defer func() {
		if err == nil {
			err = middleware.NextOnWebSocketFrame(session, messageType, message)
		}
	}()
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
	return
}
func (middleware *WsUpstreamMiddleware) OnRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	defer func() {
		if err == nil {
			err = middleware.NextOnJSONRpcRequest(session)
		}
	}()
	connSession := session.Conn
	rpcRequest := session.Request
	rpcRequestBytes := session.RequestBytes

	// create response future before to use in ProcessRpcRequest
	session.RpcResponseFutureChan = make(chan *rpc.JSONRpcResponse, 1)
	if connSession.SelectedUpstreamTarget != nil {
		session.TargetServer = *connSession.SelectedUpstreamTarget
	}

	connSession.RpcRequestsDispatchChannel <- &rpc.RpcRequestDispatchData{
		Type: rpc.RPC_REQUEST_CHANGE_TYPE_ADD_REQUEST,
		Data: session,
	}

	middleware.sendRequestToTargetConn(connSession, websocket.TextMessage, rpcRequestBytes, rpcRequest, session.RpcResponseFutureChan)
	return
}
func (middleware *WsUpstreamMiddleware) OnRpcResponse(session *rpc.JSONRpcRequestSession) (err error) {
	defer func() {
		if err == nil {
			err = middleware.NextOnJSONRpcResponse(session)
		}
	}()
	connSession := session.Conn
	defer func() {
		// notify connSession this rpc response is end
		connSession.RpcRequestsDispatchChannel <- &rpc.RpcRequestDispatchData{
			Type: rpc.RPC_REQUEST_CHANGE_TYPE_ADD_RESPONSE,
			Data: session,
		}
	}()
	responseBytes, jsonErr := json.Marshal(session.Response)
	if jsonErr == nil {
		log.Debugf("upstream response: %s", string(responseBytes))
	}
	return
}

func (middleware *WsUpstreamMiddleware) ProcessRpcRequest(session *rpc.JSONRpcRequestSession) (err error) {
	defer func() {
		if err == nil {
			err = middleware.NextProcessJSONRpcRequest(session)
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
	case <-time.After(middleware.options.upstreamTimeout):
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
