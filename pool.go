package wytcp

import (
	"sync"
)

var mutex sync.RWMutex

type ConnPool struct {
	sync.RWMutex
	items map[interface{}]*Item
}

type Item struct {
	key  interface{}
	conn *Conn
}

// ConnPool connection pool
func NewPool() *ConnPool {

	return &ConnPool{
		items: make(map[interface{}]*Item),
	}
}

func (pool *ConnPool) GetConn(key interface{}) (*Conn, bool) {
	if key == nil {
		return nil, false
	}

	pool.RLock()
	i, ok := pool.items[key]
	pool.RUnlock()
	if ok {
		return i.conn, true
	}
	return nil, false
}

func (pool *ConnPool) JoinConn(key interface{}, c *Conn) bool {
	if c == nil || key == nil {
		return false
	}

	i := &Item{key: key, conn: c}
	pool.Lock()
	pool.items[key] = i
	pool.Unlock()
	return true
}

// ExistConn
func (pool *ConnPool) ExistConn(key interface{}) bool {

	pool.RLock()

	_, ok := pool.items[key]
	pool.RUnlock()
	return ok
}

// DeleteConn delete connections by key ,such as userID
func (pool *ConnPool) DeleteConn(key interface{}) (*Conn, bool) {

	pool.Lock()
	i, ok := pool.items[key]
	if !ok {
		pool.Unlock()
		return nil, false
	}
	delete(pool.items, key)
	pool.Unlock()
	return i.conn, ok
}

// ConnCount Get connections count
func (pool *ConnPool) ConnCount() int {

	pool.RLock()
	count := len(pool.items)
	pool.RUnlock()

	return count
}
