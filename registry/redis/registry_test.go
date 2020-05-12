package redis

import (
	"context"
	"github.com/zoowii/jsonrpc_proxygo/registry"
	"net/url"
	"testing"
)

var (
	redisEndpoint = "" // "127.0.0.1"
	redisDb       = 5
)

func TestRedisRegistry_RegisterService(t *testing.T) {
	if len(redisEndpoint) < 1 {
		return
	}
	r := NewRedisRegistry()
	defer r.Close()
	err := r.Init(RedisEndpoint(redisEndpoint), RedisDatabase(redisDb))
	if err != nil {
		t.Error(err)
		return
	}
	println("registry inited")

	err = r.RegisterService(&registry.Service{
		Name: "test1",
		Url:  "test1url",
	})
	if err != nil {
		t.Error(err)
		return
	}
	println("registered a service")
	services, err := r.ListServices()
	if err != nil {
		t.Error(err)
		return
	}
	println("services", services)

}

func TestRedisRegistry_Watch(t *testing.T) {
	if len(redisEndpoint) < 1 {
		return
	}
	r := NewRedisRegistry()
	defer r.Close()
	err := r.Init(RedisEndpoint(redisEndpoint), RedisDatabase(redisDb))
	if err != nil {
		t.Error(err)
		return
	}
	println("registry inited")

	watcher, err := r.Watch()
	if err != nil {
		t.Error(err)
		return
	}
	go func() {
		ctx := context.Background()
		c := watcher.C()
	forLoop:
		for {
			select {
			case <-ctx.Done():
				break forLoop
			case event := <-c:
				if event == nil {
					break forLoop
				}
				println("watcher receive event", event.String())
			}
		}
	}()

	service1 := &registry.Service{
		Name: "test1",
		Url:  "test1url",
	}
	err = r.RegisterService(service1)
	if err != nil {
		t.Error(err)
		return
	}
	println("registered a service")
	services, err := r.ListServices()
	if err != nil {
		t.Error(err)
		return
	}
	println("services", services)
	err = r.DeregisterService(service1)
	if err != nil {
		t.Error(err)
		return
	}
	println("deregister a service done")
	services, err = r.ListServices()
	if err != nil {
		t.Error(err)
		return
	}
	println("services after deregister", services)
}

func TestRegistryUrl(t *testing.T) {
	registryUrl := "redis://127.0.0.1:6039/1"
	uri, err := url.ParseRequestURI(registryUrl)
	if err != nil {
		t.Error(err)
		return
	}
	if uri.Scheme != "redis" {
		t.Error("scheme error")
		return
	}
	if uri.Host != "127.0.0.1:6039" {
		t.Error("uri host error")
		return
	}
	println("path", uri.Path)
	if uri.Path != "/1" {
		t.Error("uri path error")
		return
	}
}
