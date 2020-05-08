package consul

import (
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/registry"
)

type consulRegistry struct {
	mOptions *consulRegistryOptions
}

func NewConsulRegistry() *consulRegistry {
	return &consulRegistry{
		mOptions: newConsulRegistryOptions(),
	}
}

func (r *consulRegistry) Init(options ...common.Option) error {
	mOptions := r.mOptions
	for _, o := range options {
		o(mOptions)
	}
	// TODO: connect to consul
	return nil
}

func (r *consulRegistry) RegisterService(service *registry.Service) error {
	return nil // TODO
}

func (r *consulRegistry) DeregisterService(service *registry.Service) error {
	return nil // TODO
}

func (r *consulRegistry) ListServices() ([]*registry.Service, error) {
	return nil, nil // TODO
}

func (r *consulRegistry) Watch() (*registry.Watcher, error) {
	return nil, nil // TODO
}

func (r *consulRegistry) Close() error {
	return nil // TODO
}

func (r *consulRegistry) String() string {
	return "consul-registry"
}
