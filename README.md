# wytcp
A non-blocking TCP server

参考 example 中 pingpong 实现简单的即可实现自定义协议格式的非阻塞 TCP 服务器




# Install
```
go get -u github.com/naxiemolv/wytcp  
```


每一个连接拥有三个goroutine


1.接收 recvChan


* 收到的数据打包装进 recvChan
 
 
2.处理 handleChan


* 处理 recvChan 中的数据包并写入 writeChan
 
 
3.发送 writeChan

* 按顺序发送数据包


# Use

```
type Callback struct {
}

type Protocol struct {
}

func main() {
    cfg := &wytcp.Conf{
		Port:         8888, // server port
		SendChanSize: 50,   // send chan size 
		ReceiveChanSize: 50,  // receive chan size 
	}
	s, err := wytcp.CreateTCPServer(cfg, &Callback{}, &Protocol{})
	if err != nil {
		fmt.Println(err)
		return
	}

	s.Start()
}

func (callback *Callback) Error(c *wytcp.Conn, d wytcp.DataPkg, errCode int) {
}

func (callback *Callback) Connect(c *wytcp.Conn) bool {
	return true
}
func (callback *Callback) Close(c *wytcp.Conn) {
    
}

func (callback *Callback) Receive(c *wytcp.Conn, d wytcp.DataPkg) bool {
    return true
}
```
