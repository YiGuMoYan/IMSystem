package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnLineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播 channel
	Message chan string
}

// NewServer 创建一个Server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnLineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// ListenMessager 监听 Message 广播消息的 goroutin，一旦有消息发送给全部的在线 User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		// 给每个用户发送消息
		this.mapLock.Lock()
		for _, cil := range this.OnLineMap {
			cil.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// BroadCast 广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s]:%s", user.Name, msg)
	this.Message <- sendMsg
}

// Handler 当前链接的业务
func (this *Server) Handler(conn net.Conn) {
	// 用户上线
	user := NewUser(conn, this)

	user.Online()

	// 接收用户传递的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
			}

			// 提取用户消息，去除 \n
			msg := string(buf[:n-1])

			// 将得到的消息进行广播
			user.DoMessage(msg)
		}
	}()
	// 当前handle阻塞
	select {}

}

// Start 启动服务器
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer listener.Close()

	go this.ListenMessager()

	// accept
	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// go handler
		go this.Handler(conn)
	}
}
