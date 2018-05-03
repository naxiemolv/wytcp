package main

import (
	"github.com/naxiemolv/wytcp"
	"fmt"

	"io"
	"encoding/binary"
	"errors"
	"bytes"
)

func main() {
	cfg := &wytcp.Conf{
		Port:         8888,
		SendChanSize: 20,
		ReceiveChanSize: 20,
	}
	s, err := wytcp.CreateTCPServer(cfg, &Callback{}, &Protocol{})
	if err != nil {
		fmt.Println(err)
		return
	}

	s.Start()
}

type Callback struct {
}

type Protocol struct {
}

// Error callback of TCP inner error
func (callback *Callback) Error(c *wytcp.Conn, d wytcp.DataPkg, errCode int) {
	fmt.Println("data error ,error type:", errCode)
}

// Connect callback of connect event , return false to close the connection
func (callback *Callback) Connect(c *wytcp.Conn) bool {
	fmt.Printf("connect. IP:%s\n", c.RemoteAddr())
	return true
}

// Close callback of close event
func (callback *Callback) Close(c *wytcp.Conn) {
	fmt.Printf("close. IP:%s\n", c.RemoteAddr())
}


// Receive callback of receive event , return false to close the connection
func (callback *Callback) Receive(c *wytcp.Conn, d wytcp.DataPkg) bool {
	pkg := d.(*MyStruct)
	fmt.Printf("receive. data len:%d\n", pkg.len)

	c.Write(d,0)

	return true
}

// Deserialization []byte to your struct
func (p *Protocol) Deserialization(c *wytcp.Conn) (wytcp.DataPkg, error) {

	// implement your protocol here
	// for example
	// The first 4 bytes of the data packet is the length
	// 00 00 00 01 30 , length is 1, data is 0x30

	splitSize := 1024
	totalLength := 0
	maxPackageSize := 1024 * 1024 * 10 // 10 MB

	lenBuff := make([]byte, 4)
	if _, err := io.ReadFull(c.RawConn, lenBuff); err != nil {
		return nil, err
	}
	// Big-Endian mode is generally used in network transmission

	if totalLength = int(binary.BigEndian.Uint32(lenBuff)); totalLength > maxPackageSize {
		return nil, errors.New("data length exceeds limit")
	}

	lastLength := totalLength

	b := new(bytes.Buffer)

	buff := make([]byte, splitSize)

	n := 0
	var err error

	for lastLength > 0 {
		if lastLength > splitSize {
			n, err = io.ReadFull(c.RawConn, buff)

		} else {
			buff = make([]byte, lastLength)
			n, err = io.ReadFull(c.RawConn, buff)
		}

		if err != nil {
			return nil, err
		}

		lastLength = totalLength - n
		b.Write(buff)
	}

	if totalLength == b.Len() {
		return toMyStruct(b.Bytes()), nil
	}

	return nil, errors.New("read error")
}

// your struct to []byte
func (s *MyStruct) Serialize() []byte {
	return s.data
}

type MyStruct struct {
	len  int
	data []byte
}

func toMyStruct(buff []byte) *MyStruct {

	p := &MyStruct{}
	p.len = len(buff)
	p.data = buff

	return p
}
