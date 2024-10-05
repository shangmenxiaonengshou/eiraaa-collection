package logger

import (
	"fmt"
	"log"
)

// Logger 是一个用于记录错误日志的结构体
func LogError(err error) {
	fmt.Print(err.Error())
	log.Printf("[ERROR]:  %s", err.Error())
}

func LogEvent(event string) {
	fmt.Print(event)
	log.Printf("[INFO]:  %s", event)
}
