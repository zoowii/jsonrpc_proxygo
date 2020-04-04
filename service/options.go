package service

import (
	"github.com/zoowii/jsonrpc_proxygo/common"
)

type wsServiceConnPoolOptions struct {
	maxConnEachBackend  int // 到每个后端的连接池的最大连接数
	initConnEachBackend int // 到每个后端的连接池的初始连接数

	afterConnCreated func(conn ServiceConn) error
}

func MaxConnEachBackend(maxConn int) common.Option {
	return func(options common.Options) {
		pOptions := options.(*wsServiceConnPoolOptions)
		if maxConn > 0 {
			pOptions.maxConnEachBackend = maxConn
		}
	}
}

func InitConnEachBackend(initConn int) common.Option {
	return func(options common.Options) {
		pOptions := options.(*wsServiceConnPoolOptions)
		if initConn >= 0 {
			pOptions.initConnEachBackend = initConn
		}
	}
}

func AfterConnCreated(callback func(conn ServiceConn) error) common.Option {
	return func(options common.Options) {
		pOptions := options.(*wsServiceConnPoolOptions)
		pOptions.afterConnCreated = callback
	}
}
