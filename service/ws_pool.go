package service

import (
	"github.com/gorilla/websocket"
	"github.com/zoowii/jsonrpc_proxygo/common"
	"github.com/zoowii/jsonrpc_proxygo/pool"
	"sync"
)

type websocketServiceConnPool struct {
	options *wsServiceConnPoolOptions

	// 每个backend service都有一个pool.ConnPool对象. pool初始大小0，
	connPools    map[string]pool.ConnPool
	connSessions *sync.Map // 如果是有状态会话，sessionId => ServiceConn的映射
}

func NewWebsocketServiceConnPool() ServiceConnPool {
	return &websocketServiceConnPool{
		connPools:    make(map[string]pool.ConnPool),
		connSessions: &sync.Map{},
	}
}

func (pool *websocketServiceConnPool) Init(argOptions ...common.Option) error {
	options := &wsServiceConnPoolOptions{
		maxConnEachBackend:  10,
		initConnEachBackend: 0,
	}
	for _, o := range argOptions {
		o(options)
	}
	pool.options = options
	return nil
}

func (p *websocketServiceConnPool) getOrCreateConnPool(targetServiceKey string) (result pool.ConnPool, err error) {
	result, ok := p.connPools[targetServiceKey]
	if ok {
		return
	}
	connCreator := func() (conn pool.Poolable, err error) {
		// 创建到backend的连接
		// targetServiceKey 就是目标endpoint
		targetEndpoint := targetServiceKey
		c, _, err := websocket.DefaultDialer.Dial(targetEndpoint, nil)
		if err != nil {
			return
		}
		serviceConn := &WebsocketServiceConn{
			Stateful:     true,
			Endpoint:     targetEndpoint,
			ServiceKey:   targetServiceKey,
			Conn:         c,
			PoolableConn: nil,
		}
		if p.options.afterConnCreated != nil {
			err = p.options.afterConnCreated(serviceConn)
			if err != nil {
				return
			}
		}
		conn = serviceConn
		return
	}
	newConnPool, err := pool.NewConnPool(p.options.maxConnEachBackend, p.options.initConnEachBackend, connCreator)
	if err != nil {
		return
	}
	p.connPools[targetServiceKey] = newConnPool
	result = newConnPool
	return
}

func getRealConn(connWrap *pool.PoolableProxy) ServiceConn {
	conn, ok := connWrap.Real().(ServiceConn)
	if !ok {
		panic("invalid real conn type in ws pool")
	}
	return conn
}

func (p *websocketServiceConnPool) GetStatelessConn(targetServiceKey string) (conn ServiceConn, err error) {
	connPool, err := p.getOrCreateConnPool(targetServiceKey)
	if err != nil {
		return
	}
	connWrap, err := connPool.Get()
	if err != nil {
		return
	}
	conn = getRealConn(connWrap)
	conn.SetPoolableConn(connWrap)
	return
}

func (p *websocketServiceConnPool) ReleaseStatelessConn(conn ServiceConn) (err error) {
	connPool, err := p.getOrCreateConnPool(conn.GetServiceKey())
	if err != nil {
		return
	}
	err = connPool.GiveBack(conn.GetPoolableConn())
	return
}

func (p *websocketServiceConnPool) GetStatefulConn(targetServiceKey string,
	sessionId string, reuse bool) (conn ServiceConn, err error) {
	sessionConn, ok := p.connSessions.Load(sessionId)
	if ok {
		conn = sessionConn.(ServiceConn)
		return
	}
	connPool, err := p.getOrCreateConnPool(targetServiceKey)
	if err != nil {
		return
	}
	connWrap, err := connPool.Get()
	if err != nil {
		return
	}
	conn = getRealConn(connWrap)
	conn.SetPoolableConn(connWrap)
	conn.SetSessionId(sessionId)
	p.connSessions.Store(sessionId, conn)
	return
}

func (p *websocketServiceConnPool) ReleaseStatefulConn(conn ServiceConn) (err error) {
	sessionId := conn.GetSessionId()
	connPool, err := p.getOrCreateConnPool(conn.GetServiceKey())
	if err != nil {
		return
	}
	// TODO: 有状态连接不仅要归还还要关闭连接，因为不能被复用了
	err = connPool.GiveBack(conn.GetPoolableConn())
	if err != nil {
		return
	}
	p.connSessions.Delete(sessionId)
	return
}

func (p *websocketServiceConnPool) Shutdown() error {
	var err error
	for _, connPool := range p.connPools {
		// TODO: 关闭连接池中各连接
		err1 := connPool.Close()
		if err1 != nil && err == nil {
			err = err1
		}
	}
	p.connPools = nil
	return err
}
