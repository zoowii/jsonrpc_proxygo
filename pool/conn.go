package pool

import (
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Poolable: anything that can used in ConnPool
type Poolable interface {
	io.Closer
}

type PoolableProxy struct {
	pool ConnPool
	real Poolable
}

func (proxy *PoolableProxy) Real() Poolable {
	return proxy.real
}

func (conn *PoolableProxy) Close() error {
	return conn.pool.GiveBack(conn.real)
}

// ConnPool: pool of connection
type ConnPool interface {
	io.Closer
	// GetOrWait will return/create available conn or wait to have available one
	GetOrWait(maxWaitTime time.Duration) (*PoolableProxy, error)
	// Get will return/create available conn or return error if max limit exceed. This method not block
	Get() (*PoolableProxy, error)
	// GiveBack will give the {conn} back to the pool as reusable
	GiveBack(conn Poolable) error
	// close the under conn and give it back to pool. maybe called when the under connection is closed before
	Remove(conn Poolable) error
}

type UpstreamConnFactory func() (Poolable, error)

type connPool struct {
	poolDataLock       sync.RWMutex
	queueSize          int32 // count of all connections managed by the pool now
	max                int
	initSize           int
	availableInstances chan Poolable
	factory            UpstreamConnFactory
}

func NewConnPool(max, initSize int, factory UpstreamConnFactory) (pool ConnPool, err error) {
	realPool := &connPool{
		queueSize:          0,
		max:                max,
		initSize:           initSize,
		availableInstances: make(chan Poolable, max*10),
		factory:            factory,
	}
	if initSize > 0 {
		err = realPool.createConnections(initSize)
		if err != nil {
			return
		}
	}
	pool = realPool
	return
}

func (pool *connPool) atomicAddQueueSize(delta int32) {
	changed := false
	for !changed {
		if delta < 0 && pool.queueSize <= 0 {
			return
		}
		changed = atomic.CompareAndSwapInt32(&pool.queueSize, pool.queueSize, pool.queueSize+delta)
	}
}

func (pool *connPool) createConnections(count int) (err error) {
	for i := 0; i < count; i++ {
		conn, createErr := pool.factory()
		if createErr != nil {
			err = createErr
			return
		}
		pool.atomicAddQueueSize(1)
		pool.availableInstances <- conn
	}
	return
}

var (
	ErrPoolMaxSizeExceed = errors.New("pool max size exceed")
	ErrAcquireTimeout    = errors.New("pool acquire timeout")
)

func (pool *connPool) wrapConn(conn Poolable) *PoolableProxy {
	return &PoolableProxy{
		pool: pool,
		real: conn,
	}
}

func (pool *connPool) GetOrWait(maxWaitTime time.Duration) (result *PoolableProxy, err error) {
	select {
	case conn := <-pool.availableInstances:
		if conn == nil {
			result = nil
			err = os.ErrClosed
			return
		}
		result = pool.wrapConn(conn)
		return
	case <-time.After(maxWaitTime):
		err = ErrAcquireTimeout
		return
	}
}

func (pool *connPool) Get() (result *PoolableProxy, err error) {
	select {
	case conn := <-pool.availableInstances:
		if conn == nil {
			result = nil
			err = os.ErrClosed
			return
		}
		result = pool.wrapConn(conn)
		return
	default:
		if pool.queueSize >= int32(pool.max) {
			err = ErrPoolMaxSizeExceed
			return
		}
		err = pool.createConnections(1)
		if err != nil {
			return
		}
		result, err = pool.Get()
		return
	}
}

func (pool *connPool) GiveBack(conn Poolable) (err error) {
	if conn == nil {
		return
	}
	poolConnProxy, ok := conn.(*PoolableProxy)
	if !ok {
		err = errors.New("invalid connection type for this pool")
		return
	}
	if pool.queueSize > int32(pool.max) {
		err = poolConnProxy.real.Close()
		pool.atomicAddQueueSize(-1)
		return err
	}
	pool.poolDataLock.RLock()
	defer pool.poolDataLock.RUnlock()

	availableInstances := pool.availableInstances
	if availableInstances == nil {
		err = errors.New("pool closed before")
		return
	}
	availableInstances <- poolConnProxy.real
	return
}

func (pool *connPool) Close() error {
	pool.poolDataLock.Lock()
	defer pool.poolDataLock.Unlock()

	close(pool.availableInstances)
	pool.availableInstances = nil
	pool.max = 0
	return nil
}

func (pool *connPool) Remove(conn Poolable) (err error) {
	if conn == nil {
		return
	}
	_, ok := conn.(*PoolableProxy)
	if !ok {
		err = errors.New("invalid connection type for this pool")
		return
	}
	if pool.queueSize > 0 {
		pool.atomicAddQueueSize(-1)
	}
	return
}
