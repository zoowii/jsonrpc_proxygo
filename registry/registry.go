package registry

import (
	"github.com/zoowii/jsonrpc_proxygo/common"
)

type Registry interface {
	Init(...common.Option) error
	RegisterService() error
	DeregisterService() error
	ListServices() error
	Watch() (Watcher, error)
	String() string
}
