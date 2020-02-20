package pool

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var mockConnIdGen = 0

type mockConn struct {
	data int
}

func (conn *mockConn) Close() error {
	return nil
}

func mockConnFactory() (result Poolable, err error) {
	id := mockConnIdGen
	mockConnIdGen++
	fmt.Printf("mock conn #%d created\n", id)
	return &mockConn{
		data: id,
	}, nil
}

func getMockConn(conn *PoolableProxy) *mockConn {
	real, _ := conn.Real().(*mockConn)
	return real
}

func TestUpstreamConnPool(t *testing.T) {
	cp, err := NewConnPool(2, 0, mockConnFactory)
	assert.True(t, err == nil)
	conn1, err := cp.Get()
	assert.True(t, err == nil)
	mockConn1 := getMockConn(conn1)
	fmt.Printf("mock conn1 data %d\n", mockConn1.data)
	conn2, err := cp.Get()
	assert.True(t, err == nil)
	mockConn2 := getMockConn(conn2)
	fmt.Printf("mock conn2 data %d\n", mockConn2.data)
	assert.True(t, mockConn1.data != mockConn2.data)

	_, err = cp.Get()
	assert.True(t, err != nil)
	println(err.Error())
	assert.True(t, ErrPoolMaxSizeExceed == err)

	cp.GiveBack(conn1)
	conn3, err := cp.Get()
	assert.True(t, err == nil)
	mockConn3 := getMockConn(conn3)
	fmt.Printf("mock conn3 data %d\n", mockConn3.data)
	assert.True(t, mockConn3.data == mockConn1.data)
}
