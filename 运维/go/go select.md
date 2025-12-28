# Go select

go 中的select 是等待操作系统的系统调用，可以让goroutine 同时等待多个channel 的可读或者可写状态，在多个文件或者 Channel 状态改变之前，select 会一直阻塞当前线程或者 Goroutine。

### select 结构体


select 在 Go 语言的源代码中不存在对应的结构体，但是我们使用 runtime.scase 结构体表示 select 控制结构中的 case：

```go
type scase struct {
	c    *hchan         // chan
	elem unsafe.Pointer // data element
}
```

因为非默认的 case 中都与 Channel 的发送和接收有关，所以 runtime.scase 结构体中也包含一个 runtime.hchan 类型的字段存储 case 中使用的 Channel。




### select 示例
select 中的等待是并发等待，任意一个channel 触发都会让switch 执行对应的代码

示例代码
```go
package main

import (
	"fmt"
	"time"
)

func main() {
	animals := make(chan string)
	quit := make(chan string)
	go testswitch(animals, quit)
	time.Sleep(time.Second)
	SendMessage(animals, quit)
	time.Sleep(time.Second)

}
func testswitch(animals, num chan string) {
	var name string
	var numstring string
	for {
		select {
		case name = <-animals:
			fmt.Println(name)
		case numstring = <-num:
			fmt.Println(numstring)

		}
	}
}
func SendMessage(animals, quit chan string) {
	animals <- "pig"
	time.Sleep(time.Second)
	animals <- "cow"
	quit <- "1"
	time.Sleep(time.Second)
	animals <- "people"
	quit <- "114514"
	time.Sleep(time.Second)
	quit <- "6"

}


```
输出
```
pig
cow
1
people
114514
6

```

select 如果在一次查看中 发现多个cannel 都触发了会随机选择一个执行

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	animals := make(chan string)
	quit := make(chan string)
	go SendMessage(animals, quit)
	testswitch(animals, quit)
}
func testswitch(animals, num chan string) {
	for {
		select {
		case <-animals:
			fmt.Println("case1")
		case <-animals:
			fmt.Println("case2")
		}
	}
}
func SendMessage(animals, quit chan string) {
	for range time.Tick(1 * time.Second) {
		animals <- "1"
	}
}


```
输出结果
```
case2
case1
case1
case2
case2
case1
case2
case1
case1
case1
case1
case1
case2
case2
....
```

当select 中不存在case和default的时候会直接阻塞 相当于执行runtime.block

当有default 语句的时候基本就是非主阻塞的