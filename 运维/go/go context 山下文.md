# go Context

### 什么是Context
Context 是一个接口类型,主要用于go中并发编程的控制

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool) // 截止时间
	Done() <-chan struct{}  //携程结束
	Err() error  //错误信息
	Value(key any) any // 携程数据
}

```
#### context 主要功能一：**数据传递**
简单例子：

```go

package main

import (
	"context"
	"fmt"
)

type UserInfo struct {
	Name string
	Age  int
}

func main() {
	//生成一个上下文
	ctx := context.Background()
	// 传递数据
	ctx = context.WithValue(ctx, "name", "eiraaa")
	ctx = context.WithValue(ctx, "name2", UserInfo{
		Name: "eiraaa2",
		Age:  18,
	})
	GetUser(ctx)

}

func GetUser(ctx context.Context) {
	fmt.Println(ctx.Value("name"))
	fmt.Println(ctx.Value("name2"))
	//断言
	fmt.Println(ctx.Value("name2").(UserInfo).Age)

}




```
输出结果

```txt
eiraaa
{eiraaa2 18}
18

```
#### Context 主要功能二: **取消**
context 中有三种取消方式
- 取消携程 WithCancel
- 超时取消 
- 截止时间取消

取消携程 WithCancel
通过生成一个带 cancel 函数的 ctx 在调用cancel 的时候会向ctx.Done这个chan 里面写入，通过检测这个chan 可以实现结束携程

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type UserInfo struct {
	Name string
	Age  int
}

var wait = sync.WaitGroup{}

func main() {
	wait.Add(3)
	fmt.Printf("并发任务开始：%s\n", time.Now())
	ctx1, cancelfunc1 := context.WithCancel(context.Background())
	ctx2, cancelfunc2 := context.WithCancel(context.Background())
	ctx3, cancelfunc3 := context.WithCancel(context.Background())
	go func() {
		WastTime(1, ctx1)

	}()
	go func() {
		WastTime(2, ctx2)

	}()
	go func() {
		WastTime(3, ctx3)
	}()

	go func() {
		// 取消携程2
		time.Sleep(1 * time.Second)
		cancelfunc1()
		time.Sleep(1 * time.Second)
		cancelfunc2()
		time.Sleep(1 * time.Second)
		cancelfunc3()
	}()
	wait.Wait()
	fmt.Printf("并发任务结束：%s\n", time.Now())
}

// 耗时操作
func WastTime(num int, ctx context.Context) {
	time1 := time.Now()
	defer func() {
		fmt.Printf("任务%d 经过%s 执行完成\n", num, (time.Since(time1)))
		wait.Done()
	}()

	go func() {
		select {
		case <-ctx.Done(): //执行就取消就取消
			{
				fmt.Printf("任务%d 经过%s 执行完成\n", num, (time.Since(time1)))
				wait.Done()
			}
		}
	}()

	fmt.Printf("任务%d 开始执行\n", num)
	time.Sleep(5 * time.Second)

}


```

```txt 
并发任务开始：2025-12-28 00:29:53.918588165 
任务1 开始执行
任务2 开始执行
任务3 开始执行
任务1 经过1.000370076s 执行完成
任务2 经过2.00094081s 执行完成
任务3 经过3.001245643s 执行完成
并发任务结束：2025-12-28 00:29:56.920040418 

```


截止时间取消 WithDeadline
生产一个带deadline 的ctx 到点会自动执向ctx.Done 里面写入

```go

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type UserInfo struct {
	Name string
	Age  int
}

var wait = sync.WaitGroup{}

func main() {
	wait.Add(3)
	fmt.Printf("并发任务开始：%s\n", time.Now())
	ctx1, _ := context.WithDeadline(context.Background(), time.Now().Add(1*time.Second))
	ctx2, _ := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	ctx3, _ := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	go func() {
		WastTime(1, ctx1)

	}()
	go func() {
		WastTime(2, ctx2)

	}()
	go func() {
		WastTime(3, ctx3)
	}()
	wait.Wait()
	fmt.Printf("并发任务结束：%s\n", time.Now())
}

// 耗时操作
func WastTime(num int, ctx context.Context) {
	time1 := time.Now()
	defer func() {
		fmt.Printf("任务%d 经过%s 执行完成\n", num, (time.Since(time1)))
		wait.Done()
	}()

	go func() {
		select {
		case <-ctx.Done(): //执行就取消就取消
			{
				fmt.Printf("任务%d 经过%s 执行完成\n", num, (time.Since(time1)))
				wait.Done()
			}
		}
	}()

	fmt.Printf("任务%d 开始执行\n", num)
	time.Sleep(5 * time.Second)

}



```

超时时间取消  WithTimeout
和截止时间基本一样，只是截止时间设定的是具体到的点，超时时间是具体等待秒数这里不做展示
