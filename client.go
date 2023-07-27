package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	// 当前客户的模式
	flag int
}

// 创建客户端
func newClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       -1,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}

	client.conn = conn

	return client
}

// Menu 显示菜单
func (client *Client) Menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("请输入合法数字")
		return false
	}
}

// PublicChat 公聊模式
func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>>>>> 请输入聊天内容，exit表示退出")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>>>>> 请输入聊天内容，exit表示退出")
		fmt.Scanln(&chatMsg)
	}
}

// SelectUser 查询在线用户
func (client *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

// PrivateChat 私聊模式
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUser()
	fmt.Println(">>>>>>>> 请输入用户名，exit退出：")
	fmt.Scanln(&remoteName)

	if remoteName != "exit" {
		fmt.Println(">>>>>>>> 请输入消息内容，exit退出：")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := fmt.Sprintf("to|%s|%s\n\n", remoteName, chatMsg)
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>>>> 请输入消息内容，exit退出：")
			fmt.Scanln(&chatMsg)

		}
	}
}

// UpdateName 更新用户昵称
func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>>>> 请输入用户名：")
	fmt.Scanln(&client.Name)
	sendMsg := fmt.Sprintf("rename|%s\n", client.Name)
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write:", err)
		return false
	}
	return true
}

// Run 持续运行
func (clint *Client) Run() {
	for clint.flag != 0 {
		for !clint.Menu() {
		}
		switch clint.flag {
		case 1:
			// 公聊模式
			clint.PublicChat()
			break
		case 2:
			// 私聊模式
			clint.PrivateChat()
			break
		case 3:
			// 更新用户名
			clint.UpdateName()
			break
		}
	}
}

// DealResponse 处理 server 返回消息，直接显示标准输出
func (client *Client) DealResponse() {
	// 一旦 client,conn 有数据，就 copy 到 stdout 标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认为127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认为8888)")
}

func main() {
	flag.Parse()

	client := newClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>> 连接服务器失败...")
		return
	}

	// 单独开启一个 goroutine 来处理新消息
	go client.DealResponse()

	fmt.Println(">>>>>>>> 连接服务器成功...")

	// 启动客户端业务
	client.Run()
}
