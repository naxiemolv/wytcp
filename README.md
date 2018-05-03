# wytcp
A non-blocking TCP server
一个非阻塞的 TCP 服务器

Install

go get -u github.com/naxiemolv/wytcp

每一个连接拥有三个goroutine


1.接收 recvChan


* 收到的数据打包装进 recvChan
 
 
2.处理 handleChan


* 处理 recvChan 中的数据包并写入 writeChan
 
 
3.发送 writeChan

* 按顺序发送数据包
