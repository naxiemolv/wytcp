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
func GetConnPool() *ConnPool {

	if pool == nil {
		pool = &ConnPool{
			items: make(map[interface{}]*Item),
		}
	}

	return pool
}

func JoinConn(key interface{}, c *Conn) bool {
	if c == nil || key == nil {
		return false
	}
	p := GetConnPool()
	i := &Item{key: key, conn: c}
	p.Lock()
	p.items[key] = i
	p.Unlock()
	return true
}

// ExistConn
func ExistConn(key interface{}) bool {
	p := GetConnPool()
	p.RLock()
	defer p.RUnlock()
	_, ok := p.items[key]
	return ok
}

// DeleteConn delete connections by key ,such as userID
func DeleteConn(key interface{}) (*Conn, bool) {
	p := GetConnPool()
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
func ConnCount() int {
	p := GetConnPool()
	p.RLock()
	defer p.RUnlock()
	return len(p.items)
}
