package handler

import (
	"bytes"
	"eiradinner/internal/logger"
	"eiradinner/internal/structs"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var FileNAME = make(chan string, 5)
var CurrentDownloadFile string
var CurrentUploadFile string

func RemoveFirstOccurrence(input string, toRemove string) string {
	index := strings.Index(input, toRemove)
	if index == -1 {
		return input // 如果找不到，则返回原字符串
	}
	return input[:index] + input[index+len(toRemove):]
}

func ReceiveClientMessage(conn net.Conn) {
	// 	buffer := make([]byte, 1024*1024*2)
	// 	for {
	// 		n, _ := conn.Read(buffer)
	// 		// if err != nil {
	// 		// 	fmt.Println("读取数据时出错:", err)
	// 		// 	return
	// 		// }
	// 		if n > 0 {
	// 			HandleclientMesseage(conn, buffer[:n]) // 执行相应的命令
	// 		}
	// 	}
	// }
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
			logger.LogEvent(fmt.Sprintf("缓冲区长度:%s", buffer.Len()))
			// 处理接收到的数据
			HandleclientMesseage(conn, buffer.Bytes())
			fmt.Print("执行结束，清空缓冲区")

			buffer.Reset() // 清空缓冲区
			fmt.Println("缓冲区长度:", buffer.Len())
			logger.LogEvent(fmt.Sprintf("缓冲区长度:%s", buffer.Len()))
		}
	}
}

func HandleclientMesseage(conn net.Conn, data []byte) {
	var message structs.Hartbit
	fmt.Println(string(data))
	err := json.Unmarshal(data, &message)
	if err != nil {
		fmt.Println("解码数据时出错:", err)
		return
	}
	// 根据 messageType 执行相应的操作
	switch message.MessageType {
	case 3:
		var execresult structs.ClientResponse
		json.Unmarshal(data, &execresult)
		// fmt.Println(execresult)
		utf8result, _ := convertToUTF8(execresult.Content)
		fmt.Print(utf8result)
	case 4: //从client 下载文件
		fmt.Print("开始接收文件")
		filesavepath := CurrentDownloadFile
		Resivefile(conn, data, filesavepath)

	}

}

func SendCommand(conn net.Conn, command string) {
	execommand := structs.ClientCommand{
		MessageType: 1,
		Timestamp:   time.Now().Format(time.RFC3339),
		Command:     command,
	}
	execommandmessage, _ := json.Marshal(execommand)
	conn.Write(execommandmessage)
}

func convertToUTF8(data []byte) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder() // 使用 GBK 编码
	utf8Data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(data), decoder))
	if err != nil {
		return "", err
	}
	return string(utf8Data), nil
}

func Resivefile(conn net.Conn, data []byte, filesavepath string) {
	fmt.Println(filesavepath)
	n := len(data)
	var msg structs.FileTransferMessage
	fmt.Println(string(data[:n]))
	err := json.Unmarshal(data[:n], &msg)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}
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

func Sendfile(conn net.Conn, data []byte) { //向server 端传输文件
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
