package wytcp

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ErrorTypeUnknown      = 0
	ErrorTypeSendFailed   = -1
	ErrorTypeHandleFailed = -2
	ErrorTypeWriteFailed  = -3
)

var (
	ErrConnClosed    = errors.New("conn closed")
	ErrWriteBlocking = errors.New("write blocking")
	ErrReadBlocking  = errors.New("read blocking")
)

type Conn struct {
	server       *Server
	RawConn      *net.TCPConn
	SendCount    int64 // how many package sent during this connect
	ReceiveCount int64 // how many package received during this connect
	sendChan     chan DataPkg
	receiveChan  chan DataPkg
	closeChan    chan bool
	closeOnce    sync.Once
	closeFlag    int32
	hbRequireItv time.Duration
	userData     interface{}
}

type Callback interface {
	Connect(*Conn) bool

	Close(*Conn)

	Receive(*Conn, DataPkg) bool

	Error(*Conn, DataPkg, int)
}

type Receiver interface {
	Deserialization(*Conn) (DataPkg, error)
}

type DataPkg interface {
	Serialize() []byte
}

func joinConn(conn *net.TCPConn, server *Server) {
	c := &Conn{
		server:       server,
		RawConn:      conn,
		SendCount:    0,
		ReceiveCount: 0,
		sendChan:     make(chan DataPkg, server.cfg.SendChanSize),
		receiveChan:  make(chan DataPkg, server.cfg.ReceiveChanSize),
		closeChan:    make(chan bool),
		hbRequireItv: server.cfg.HeartBeatCheckItv,
	}
	go func() {
		if c.server.callback.Connect(c) {
			go c.handleLoop()
			go c.receiveLoop()
			go c.sendLoop()

		}
	}()

}

func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		close(c.sendChan)
		close(c.receiveChan)
		c.RawConn.Close()
		c.server.callback.Close(c)
	})
}

func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}

func (c *Conn) receiveLoop() {
	defer func() {
		if e := recover(); e != nil {
			runtimeError(e)
		}
		c.Close()
	}()

	for {
		select {
		case <-c.server.exitChan:
			return
		case <-c.closeChan:
			return
		default:
		}

		d, err := c.server.receiver.Deserialization(c)
		if c.hbRequireItv > 0 {
			msg := make(chan bool)
			go heartBeating(c.RawConn, msg, c.hbRequireItv)
			go gravelChannel(msg)
		}

		if err != nil {
			return
		}
		c.receiveChan <- d
		c.ReceiveCount++
	}
}

func heartBeating(conn net.Conn, readerChannel chan bool, timeout time.Duration) {
	select {
	case fk := <-readerChannel:
		if fk {
			conn.SetDeadline(time.Now().Add(timeout))
		}

		break
	case <-time.After(time.Second * 5):
		conn.Close()
	}

}

func gravelChannel(mess chan bool) {
	mess <- true
	close(mess)
}

func (c *Conn) SetMinHeartBeatTime(duration time.Duration) {
	c.hbRequireItv = duration
}

func (c *Conn) handleLoop() {
	defer func() {
		if e := recover(); e != nil {
			runtimeError(e)
		}
		c.Close()
	}()
	for {
		select {
		case <-c.server.exitChan:
			return

		case <-c.closeChan:
			return
		case d := <-c.receiveChan:
			if c.IsClosed() {
				c.server.callback.Error(c, d, ErrorTypeHandleFailed)
				return
			}
			if !c.server.callback.Receive(c, d) {
				return
			}
		}
	}
}
func (c *Conn) sendLoop() {
	defer func() {
		if e := recover(); e != nil {
			runtimeError(e)
		}
		c.Close()
	}()

	for {
		select {
		case <-c.server.exitChan:
			return
		case <-c.closeChan:
			return
		case d := <-c.sendChan:
			if c.IsClosed() {
				c.server.callback.Error(c, d, ErrorTypeSendFailed)
				return
			}
			if _, err := c.RawConn.Write(d.Serialize()); err != nil {
				return
			}
			c.SendCount++
		}

	}
}

func (c *Conn) Write(d DataPkg, timeout time.Duration) error {
	var err error
	defer func() {
		if re := recover(); re != nil {
			runtimeError(err)
		}
		err = ErrConnClosed
	}()
	if c.IsClosed() {
		c.server.callback.Error(c, d, ErrorTypeWriteFailed)
		return ErrConnClosed
	}
	if timeout == 0 {
		select {
		case c.sendChan <- d:
			return nil

		default:
			return ErrWriteBlocking
		}

	} else {
		select {
		case c.sendChan <- d:
			return nil

		case <-c.closeChan:
			return ErrConnClosed

		case <-time.After(timeout):
			return ErrWriteBlocking
		}
	}

}

func runtimeError(e interface{}) {
	fmt.Println("[wytcp] error:", e)
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.RawConn.RemoteAddr()
}

// SetUserData set user data
func (c *Conn) SetUserData(i interface{}) {
	c.userData = i
}

// GetUserData get user data
func (c *Conn) GetUserData() interface{} {
	return c.userData
}
