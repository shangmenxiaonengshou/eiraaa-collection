package structs

import "net"

type Sessions struct {
	SessionID int
	Address   string
	Port      string
	Os        string
	Path      string
	Status    string
	Listener  string
	User      string
	Conn      net.Conn
}

type Listener struct {
	ListenerID int
	Name       string
	Addr       string
	Port       string
	Listener   net.Listener
}

type C2clientMessage struct {
	MagicNumber uint32
	MessageType uint16
	Message     []byte
}

type Hartbit struct {
	MessageType int    `json:"messagetype"`
	Timestamp   string `json:"timestamp"`
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

type FileMessage struct { //上传时是将文件放到FilePath中，下载时是将FilePath中指定的文件下载
	MessageType int    `json:"messagetype"`
	Timestamp   string `json:"timestamp"`
	FilePath    string `json:"file_name"`
}

type ClientResponse struct {
	MessageType int    `json:"messagetype"` //3
	Timestamp   string `json:"timestamp"`
	Content     []byte `json:"content"`
}

type ClientCommand struct {
	MessageType int    `json:"messagetype"` //2
	Timestamp   string `json:"timestamp"`
	Command     string `json:"command"`
}

//MessageType : 0 = hartbit, 1 = command, 2 = filedownload,4 = fileupload, 3 = cmdresponse， 5 = firenamemessage
