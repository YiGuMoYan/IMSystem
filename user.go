package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// Online 用户上线
func (this *User) Online() {
	// 用户上线，将用户添加到 OnLineMap 中
	this.server.mapLock.Lock()
	this.server.OnLineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

// Offline 用户下线
func (this *User) Offline() {
	// 用户上线，将用户添加到 OnLineMap 中
	this.server.mapLock.Lock()
	delete(this.server.OnLineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, "已下线")
}

// SendMsg 给当前用户客户端发消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// DoMessage 用户处理信息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户都有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnLineMap {
			onLineMsg := fmt.Sprintf("[%s]:在线...\n", user.Name)
			this.SendMsg(onLineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 消息格式 rename|张三
		newName := strings.Split(msg, "|")[1]

		// 判断名称是否存在
		_, ok := this.server.OnLineMap[newName]
		if ok {
			this.SendMsg("当前用户名被使用\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnLineMap, this.Name)
			this.server.OnLineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg(fmt.Sprintf("您的用户名已更新为：%s\n", newName))
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式 to|张三|消息内容
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式错误，请使用：to|用户名|消息\n")
			return
		}
		// 根据用户名，获取对方的 user 对象
		remoteUser, ok := this.server.OnLineMap[remoteName]
		if !ok {
			this.SendMsg("当前用户不存在，请使用 who 查看上线用户\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(fmt.Sprintf("[%s]私聊：%s\n", this.Name, content))
	} else {
		this.server.BroadCast(this, msg)
	}

}

// NewUser 创建一个用户 API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

// ListenMessage 监听当前 user 的 channel，一旦有消息，发送给对端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
