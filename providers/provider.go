package providers

import (
	"github.com/zoowii/jsonrpc_proxygo/rpc"
	"github.com/zoowii/jsonrpc_proxygo/utils"
)

var log = utils.GetLogger("provider")

type RpcProviderProcessor interface {
	NotifyNewConnection(connSession *rpc.ConnectionSession) error
	OnConnectionClosed(connSession *rpc.ConnectionSession) error
	OnRawRequestMessage(connSession *rpc.ConnectionSession, rpcSession *rpc.JSONRpcRequestSession,
		messageType int, message []byte) error
	OnRpcRequest(connSession *rpc.ConnectionSession, rpcSession *rpc.JSONRpcRequestSession) error
}

type RpcProvider interface {
	SetRpcProcessor(processor RpcProviderProcessor)
	ListenAndServe() error
}
