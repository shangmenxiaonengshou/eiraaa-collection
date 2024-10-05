package main

import (
	"fmt"
	"io"
	"os"
)

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
