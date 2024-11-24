package main

import (
	"bufio"
	"eiradinner/internal/databases"
	"eiradinner/internal/modules"
	"fmt"
	"io"
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
			modules.HandleGenerate(reader)
		case "exit":
			fmt.Println("退出程序")
			os.Exit(0)
		default:
			fmt.Println("无效指令，使用 help 查看帮助")
		}
	}

}

func Init_Path() error {
	// 定义需要创建的路径
	directories := []string{"db", "log"}

	for _, dir := range directories {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			if os.IsExist(err) {
				continue
			} else {
				fmt.Printf("创建目录 %s 时出错: %v\n", dir, err)
				return err
			}
		} else {
			fmt.Printf("成功创建目录: %s\n", dir)
		}
	}
	return nil
}

func Init_LogFile() (io.Writer, error) {
	// 定义需要创建的路径
	f, err := os.OpenFile("./log/eiradinner.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return f, nil

}
