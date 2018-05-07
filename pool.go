package wytcp

import (
	"sync"
)

var mutex sync.RWMutex
var pool *ConnPool

type ConnPool struct {
	sync.RWMutex
	items map[interface{}]*Item
}

type Item struct {
	sync.RWMutex
	key  interface{}
	conn *Conn
}

// ConnPool connection pool
func Pool() *ConnPool {

	if pool == nil {
		pool = &ConnPool{
			items: make(map[interface{}]*Item),
		}
	}

	return pool
}

func (pool *ConnPool) GetConn(key interface{}) (*Conn, bool) {
	if key == nil {
		return nil, false
	}
	p := Pool()
	p.RLock()
	i, ok := p.items[key]
	p.RUnlock()

	return i.conn, ok
}

func (pool *ConnPool) JoinConn(key interface{}, c *Conn) bool {
	if c == nil || key == nil {
		return false
	}
	p := Pool()
	i := &Item{key: key, conn: c}
	p.Lock()
	p.items[key] = i
	p.Unlock()
	return true
}

// ExistConn
func (pool *ConnPool) ExistConn(key interface{}) bool {
	p := Pool()
	p.RLock()
	defer p.RUnlock()
	_, ok := p.items[key]
	return ok
}

// DeleteConn delete connections by key ,such as userID
func (pool *ConnPool) DeleteConn(key interface{}) (*Conn, bool) {
	p := Pool()
	p.Lock()
	i, ok := p.items[key]
	if !ok {
		return nil, false
	}
	delete(p.items, key)
	p.Unlock()
	return i.conn, ok
}

// ConnCount Get connections count
func (pool *ConnPool) ConnCount() int {
	p := Pool()
	p.RLock()
	defer p.RUnlock()
	return len(p.items)
}
