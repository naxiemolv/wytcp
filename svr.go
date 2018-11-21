package wytcp

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type Conf struct {
	Port              uint16
	SendChanSize      int
	ReceiveChanSize   int
	AcceptTimeout     time.Duration
	HeartBeatCheckItv time.Duration
}

type Server struct {
	cfg      *Conf
	callback Callback
	receiver Receiver
	exitChan chan bool
	listener *net.TCPListener
}

// CreateTCPServer create a tcp server
func CreateTCPServer(cfg *Conf, callback Callback, receiver Receiver) (*Server, error) {
	if cfg == nil || callback == nil || receiver == nil {
		return nil, errors.New("miss parameter")
	}
	s := &Server{
		cfg:      cfg,
		callback: callback,
		receiver: receiver,
		exitChan: make(chan bool),
	}
	if cfg.Port == 0 {
		return nil, errors.New("tcp server need to bind a port")
	}
	addr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	s.listener = listener
	return s, nil
}

func (s *Server) Start() {
	defer func() {
		s.listener.Close()
	}()

	for {

		select {
		case <-s.exitChan:
			return
		default:
		}
		if s.cfg.AcceptTimeout > 0 {
			s.listener.SetDeadline(time.Now().Add(s.cfg.AcceptTimeout))
		}
		s.listener.SetDeadline(time.Now().Add(time.Second))
		c, err := s.listener.AcceptTCP()
		if err != nil {
			continue
		}

		joinConn(c, s)
	}
}
