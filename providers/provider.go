package providers

import (
	"github.com/zoowii/jsonrpc_proxygo/plugin"
	"github.com/zoowii/jsonrpc_proxygo/utils"
)

var log = utils.GetLogger("provider")

type RpcProvider interface {
	SetMiddlewareChain(chain *plugin.MiddlewareChain)
	ListenAndServe() error
}
