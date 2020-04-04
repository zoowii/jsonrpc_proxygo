package service

import "github.com/zoowii/jsonrpc_proxygo/common"

// ServiceConnPool is connection pool connected to target backend services
type ServiceConnPool interface {
	Init(...common.Option) error
	// GetStatelessConn 获取目标backend service的一个无状态连接
	GetStatelessConn(targetServiceKey string) (ServiceConn, error)
	// ReleaseStatelessConn 释放一个无状态连接到连接池
	ReleaseStatelessConn(conn ServiceConn) error
	// GetStatefulConn 获取一个有状态连接，参数{sessionId}是本会话的id
	// 同一个会话需要使用同一个有状态连接. {reuse}参数表示这个连接是否能同时给多个会话提供服务.
	// 也就是区分客户端会话和backend连接一对一，以及多对一两种情况
	GetStatefulConn(targetServiceKey string, sessionId string, reuse bool) (ServiceConn, error)
	// ReleaseStatefulConn 在手动关闭有状态连接或者连接超时没被使用时释放到连接池
	ReleaseStatefulConn(conn ServiceConn) error
	Shutdown() error
}
