package redis

import (
	"context"
	"github.com/go-redis/redis/v7"
	"github.com/zoowii/jsonrpc_proxygo/common"
)

type redisRegistryOptions struct {
	Endpoint    string
	Password    string
	Db          int
	RedisClient *redis.Client
	Context     context.Context
}

func newRedisRegistryOptions() *redisRegistryOptions {
	return &redisRegistryOptions{
		Endpoint: "127.0.0.1:6379",
		Password: "",
		Db:       0,
		Context:  context.Background(),
	}
}

func RedisEndpoint(endpoint string) common.Option {
	return func(options common.Options) {
		mOptions := options.(*redisRegistryOptions)
		mOptions.Endpoint = endpoint
	}
}

func RedisPassword(password string) common.Option {
	return func(options common.Options) {
		mOptions := options.(*redisRegistryOptions)
		mOptions.Password = password
	}
}

func RedisDatabase(db int) common.Option {
	return func(options common.Options) {
		mOptions := options.(*redisRegistryOptions)
		mOptions.Db = db
	}
}

func WithContext(ctx context.Context) common.Option {
	return func(options common.Options) {
		mOptions := options.(*redisRegistryOptions)
		mOptions.Context = ctx
	}
}
