package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/registry"
	"github.com/zoowii/jsonrpc_proxygo/utils"
	"sync"
	"time"
)

var log = utils.GetLogger("redis-registry")

type redisRegistry struct {
	lock     sync.RWMutex
	mOptions *redisRegistryOptions
	services []*registry.Service
	watchers []*registry.Watcher
}

func NewRedisRegistry() *redisRegistry {
	return &redisRegistry{
		mOptions: newRedisRegistryOptions(),
	}
}

func (r *redisRegistry) Init(options ...common.Option) (err error) {
	mOptions := r.mOptions
	for _, o := range options {
		o(mOptions)
	}
	// connect to redis
	client := redis.NewClient(&redis.Options{
		Addr:     mOptions.Endpoint,
		Password: mOptions.Password,
		DB:       mOptions.Db,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		return
	}
	log.Infof("redis registry receive redis %s", pong)
	mOptions.RedisClient = client

	// 后台轮询监听redis中各services的状态，如果掉线则发起event到watchers中
	ctx := mOptions.Context
	go func() {
		timer := time.NewTimer(1 * time.Second)
	forLoop:
		for {
		selectL:
			select {
			case <-ctx.Done():
				break forLoop
			case <-timer.C:
				// 检查目前各services状态
				var servicesCopy []*registry.Service
				func() {
					r.lock.RLock()
					defer r.lock.RUnlock()
					servicesCopy = r.services[:]
				}()
				var inactiveServices []*registry.Service
				for _, s := range servicesCopy {
					active, err := r.checkServiceStatus(s)
					if err != nil {
						log.Error("check service status error", err)
						break selectL
					}
					if !active {
						inactiveServices = append(inactiveServices, s)
					}
				}
				func() {
					r.lock.Lock()
					defer r.lock.Unlock()
					for _, s := range inactiveServices {
						idx := r.indexOfService(s)
						if idx >= 0 {
							r.services = append(r.services[:idx], r.services[idx+1:]...)
						}
					}
				}()
				for _, s := range inactiveServices {
					r.notifyWatchers(registry.NewEvent(registry.SERVICE_REMOVE, s))
				}
			}
		}
	}()

	return
}

func (r *redisRegistry) notifyWatchers(event *registry.Event) {
	var watchers []*registry.Watcher
	func() {
		r.lock.RLock()
		defer r.lock.RUnlock()
		watchers = r.watchers[:]
	}()
	for _, w := range watchers {
		w.Send(event)
	}
}

/**
 * 检查某个服务是否在线
 */
func (r *redisRegistry) checkServiceStatus(service *registry.Service) (active bool, err error) {
	client := r.mOptions.RedisClient
	if client == nil {
		err = redisRegistryNotConnectToRedisError
		return
	}
	key := serviceRedisKey(service)
	value, err := client.Get(key).Result()
	if err != nil {
		active = false
		err = nil
		return
	}
	info := &registry.Service{}
	err = json.Unmarshal([]byte(value), info)
	if err != nil {
		active = false
		err = nil
		return
	}
	if info.Name == service.Name {
		active = true
		return
	}
	active = false
	return
}

var (
	redisRegistryNotConnectToRedisError = errors.New("redis registry not connect to redis error")
)

func serviceRedisKey(service *registry.Service) string {
	return fmt.Sprintf("registry-service-%s-%s", service.Name, service.Url)
}

func (r *redisRegistry) indexOfService(service *registry.Service) int {
	for i, s := range r.services {
		if s.Name == service.Name && s.Url == service.Url {
			return i
		}
	}
	return -1
}

func (r *redisRegistry) RegisterService(service *registry.Service) (err error) {
	mOptions := r.mOptions
	if mOptions.RedisClient == nil {
		err = redisRegistryNotConnectToRedisError
		return
	}
	client := mOptions.RedisClient
	key := serviceRedisKey(service)

	func() {
		// 不存在则加入services
		r.lock.Lock()
		defer r.lock.Unlock()
		idx := r.indexOfService(service)
		if idx < 0 {
			r.services = append(r.services, service)
		}
	}()
	r.notifyWatchers(registry.NewEvent(registry.SERVICE_ADD, service))

	serviceData := utils.JsonDumpsToStringSilently(service, "")
	client.Set(key, serviceData, -1*time.Second) // 目前没设置超时不要求心跳

	return
}

func (r *redisRegistry) DeregisterService(service *registry.Service) (err error) {
	mOptions := r.mOptions
	if mOptions.RedisClient == nil {
		err = redisRegistryNotConnectToRedisError
		return
	}
	client := mOptions.RedisClient
	key := serviceRedisKey(service)

	func() {
		// 如果在services中则移除
		r.lock.Lock()
		defer r.lock.Unlock()
		idx := r.indexOfService(service)
		if idx >= 0 {
			r.services = append(r.services[:idx], r.services[idx+1:]...)
		}
	}()
	r.notifyWatchers(registry.NewEvent(registry.SERVICE_REMOVE, service))

	client.Del(key)

	return
}

func (r *redisRegistry) ListServices() ([]*registry.Service, error) {
	return r.services, nil
}

func (r *redisRegistry) Watch() (*registry.Watcher, error) {
	watcher := registry.NewWatcher()
	r.lock.Lock()
	defer r.lock.Unlock()
	r.watchers = append(r.watchers, watcher)
	return watcher, nil
}

func (r *redisRegistry) Close() (err error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	client := r.mOptions.RedisClient
	if client != nil {
		err = client.Close()
		if err != nil {
			return err
		}
		r.mOptions.RedisClient = nil
	}
	watchers := r.watchers
	for _, w := range watchers {
		w.Close()
	}
	r.watchers = nil
	return nil
}

func (r *redisRegistry) String() string {
	return "redis-registry"
}
