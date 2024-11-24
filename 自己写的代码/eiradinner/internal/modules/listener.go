package modules

import (
	"bufio"
	"eiradinner/internal/handler"
	"eiradinner/internal/structs"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var LISTENER_ID int = 0
var (
	ALL_SESSIONS  = make(map[int]*structs.Sessions) // 存储连接会话的映射
	ALL_LISTENERS = make(map[int]*structs.Listener)
	mu            sync.Mutex // 保护连接会话的互斥锁
	currentMode   string     // 当前模式
)

func HandlerListener(reader *bufio.Reader) error {
	for {
		fmt.Print("listener >")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入失败:", err)
			continue
		}
		command := strings.Fields(input)
		if len(command) == 0 {
			continue // 如果切片为空，直接跳过
		}
		commandfist := string(command[0])
		switch commandfist {
		case "list":
			ListListener()
		case "create":
			Createlistener(command)
		case "delete":
			DeleteListener(command)
		case "exit":
			return nil
		default:
			fmt.Println("无效指令，使用 help 查看帮助")

		}
	}
}

func Createlistener(args []string) {

	if len(args) != 4 {
		fmt.Println("参数错误！ 用法: create listenername ip port")
		return
	}
	listener_name := args[1]
	ip := args[2]
	port := args[3]

	for _, listener := range ALL_LISTENERS {
		if listener.Name == listener_name {
			fmt.Printf("监听器名称 %s 已存在，无法重复创建\n", listener_name)
			return
		}
	}
	// 创建listener
	listener, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		fmt.Printf("无法监听端口 %s: %v\n", port, err)
		return
	}
	fmt.Printf("create  linstener  %s\n", listener_name)
	fmt.Printf("正在监听端口 %s...\n", port)
	new_listener := &structs.Listener{ListenerID: LISTENER_ID, Name: listener_name, Addr: ip, Port: port, Listener: listener}
	LISTENER_ID += 1
	ALL_LISTENERS[LISTENER_ID] = new_listener
	go Acceptconnection(listener, port, listener_name)
}

func Acceptconnection(listener net.Listener, port string, listener_name string) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("接受连接失败:", err)
			return // 退出循环，避免继续处理关闭的连接
		}
		address := conn.RemoteAddr().String()
		mu.Lock()
		ALL_SESSIONS[SESSION_ID] = &structs.Sessions{
			SessionID: SESSION_ID,
			Address:   address,
			Port:      port,
			Conn:      conn,
			Listener:  listener_name,
		}
		mu.Unlock()
		go handleConnection(conn, SESSION_ID) // 启动一个 goroutine 处理connection
	}
}

func handleConnection(conn net.Conn, session_id int) {

	fmt.Printf("客户端 %s 连接\n", conn.RemoteAddr().String())

	// 启动一个 goroutine 读取响应
	go func() {
		for {
			handler.ReceiveClientMessage(conn)
		}
	}()
}

func DeleteListener(args []string) {

	if len(args) != 2 {
		fmt.Println("参数错误！ 用法: create listenername ip port")
		return
	}
	id, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("参数错误！ 删除的listener id 必须为数字")
		log.Println(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if listener, exists := ALL_LISTENERS[id]; exists {
		err := listener.Listener.Close()
		if err != nil {
			fmt.Printf("关闭监听器 %d 时出错: %v\n", id, err)
		} else {
			delete(ALL_LISTENERS, id)
			fmt.Printf("成功关闭监听器 %d\n", id)
		}
	} else {
		fmt.Printf("监听器 %d 不存在\n", id)
	}

}

func ListListener() {
	fmt.Printf("%-12s %-10s %-15s %-10s \n", "ID", "Name", "Listen ip", "Port")
	fmt.Println(strings.Repeat("-", 80))

	for _, listener := range ALL_LISTENERS {
		fmt.Printf("%-12d %-10s %-15s %-10s \n", listener.ListenerID, listener.Name, listener.Addr, listener.Port)
	}
}

func DownloadFile(args []string) {

}

func UploadFile(args []string) {
}
