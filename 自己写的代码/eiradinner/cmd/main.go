package main

import (
	"bufio"
	"eiradinner/internal/databases"
	"eiradinner/internal/modules"
	"fmt"
	"log"
	"os"
	"strings"
)

var Session_id int

func main() {
	Session_id = 0
	Init_Path()
	logfile, err := Init_LogFile()
	if err != nil {
		fmt.Println("初始化日志文件失败！！！")
	}
	log.SetOutput(logfile)
	_, err = databases.InitDatabase()
	if err != nil {
		log.Fatal("初始化数据库失败！！！")
	}
	log.Println("初始化数据库成功！！！")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入失败:", err)
			continue
		}
		command := strings.TrimSpace(input)
		switch command {
		case "listener":
			modules.HandlerListener(reader)
		case "session":
			modules.HandlerSession(reader)
		case "generate":
			modules.Generate_Payload()
		case "exit":
			fmt.Println("退出程序")
			os.Exit(0)
		default:
			fmt.Println("无效指令，使用 help 查看帮助")
		}
	}

}
