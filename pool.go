package wytcp

import (
	"sync"
)

var mutex sync.RWMutex

type ConnPool struct {
	sync.RWMutex
	Items map[interface{}]*Item
}

type Item struct {
	key  interface{}
	conn interface{}
}

// ConnPool connection pool
func NewPool() *ConnPool {

	return &ConnPool{
		Items: make(map[interface{}]*Item),
	}
}

func (pool *ConnPool) GetConn(key interface{}) (interface{}, bool) {
	if key == nil {
		return nil, false
	}

	pool.RLock()
	i, ok := pool.Items[key]
	pool.RUnlock()
	if ok {
		return i.conn, true
	}
	return nil, false
}

func (pool *ConnPool) JoinConn(key interface{}, c interface{}) bool {
	if c == nil || key == nil {
		return false
	}

	i := &Item{key: key, conn: c}
	pool.Lock()
	pool.Items[key] = i
	pool.Unlock()
	return true
}

// ExistConn
func (pool *ConnPool) ExistConn(key interface{}) bool {

	pool.RLock()

	_, ok := pool.Items[key]
	pool.RUnlock()
	return ok
}

// DeleteConn delete connections by key ,such as userID
func (pool *ConnPool) DeleteConn(key interface{}) (interface{}, bool) {

	pool.Lock()
	i, ok := pool.Items[key]
	if !ok {
		pool.Unlock()
		return nil, false
	}
	delete(pool.Items, key)
	pool.Unlock()
	return i.conn, ok
}

func (pool *ConnPool) QuitConn(key interface{}, conn interface{}) (bool) {
	pool.Lock()
	defer pool.Unlock()
	i, ok := pool.Items[key]
	if !ok {

		return false
	}
	if i.conn != conn {

		return false
	}

	delete(pool.Items, key)

	return true
}

// ConnCount Get connections count
func (pool *ConnPool) ConnCount() int {

	pool.RLock()
	count := len(pool.Items)
	pool.RUnlock()

	return count
}
