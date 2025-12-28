# go sync锁
sync.Mutex是Go中的互斥锁,在goroutine 并发编程中起着重要作用，互斥锁在并发过程中只能有一个goroutine 能够持有锁，适合读写都频繁，且操作时间段的场景

RWmutex 读写锁在并发过程中多个读可以并发，提升并发性能，适合读多写少，读操作耗时的场景


### 结构

```go

type Mutex struct {
	_ noCopy

	mu isync.Mutex
}

type Mutex struct {
	state int32
	sema  uint32
}

const (
    mutexLocked = 1 << iota // mutex is locked - 00000001
    mutexWoken              // 已经被唤醒 - 00000010
    mutexStarving           // 处于饥饿模式 - 00000100
    mutexWaiterShift = iota // 等待者数量移位值 = 3
    starvationThresholdNs = 1e6 // 饥饿模式阈值 = 1毫秒
)


```



+---------------+---------------+---------------+---------------+
|  等待者数量   |  饥饿标志位   |  唤醒标志位   |  锁定标志位   |
|   (29位)     |   (1位)      |   (1位)      |   (1位)      |
+---------------+---------------+---------------+---------------+
31              3              2              1              0

锁方法
```go
func (m *Mutex) Lock() {
	//快路径，直接获取未被占有的互斥锁
	if atomic.CompareAndSwapInt32(&m.state, 0, mutexLocked) {
		if race.Enabled {
			race.Acquire(unsafe.Pointer(m))
		}
		return
	}
	// 锁已被持有，复杂处理
	m.lockSlow()
}

```
- atomic.CompareAndSwapInt32：原子比较并交换操作;这块是用汇编语言来做底层实现的
- &m.state：指向 Mutex 状态字段的指针
- 0：期望的旧值（表示锁未被持有）
- mutexLocked：要设置的新值（1，表示锁已被持有）

lockSlow 方法详解
其中`new |= ` 是位设置操作，使用按位的或运算在设置特定的标志位
`<<` 是左移运算符
```
// 语法：x << n
// 含义：将 x 的二进制表示向左移动 n 位
// 效果：相当于乘以 2^n

// 示例：
1 << 0 = 1     // 1 * 2^0 = 1
1 << 1 = 2     // 1 * 2^1 = 2  
1 << 2 = 4     // 1 * 2^2 = 4
1 << 3 = 8     // 1 * 2^3 = 8


const (
    mutexLocked   = 1 << 0  // 第0位：锁定标志
    mutexWoken    = 1 << 1  // 第1位：唤醒标志  
    mutexStarving = 1 << 2  // 第2位：饥饿标志
)
```
```go

func (m *Mutex) lockSlow() {
    var waitStartTime int64  // 记录开始等待的时间，用于判断是否进入饥饿模式
    starving := false       // 当前goroutine是否处于饥饿状态
    awoke := false          // 是否已被唤醒（用于防止重复唤醒）
    iter := 0              // 自旋迭代计数器
    old := m.state         // 获取当前锁状态

    //自旋逻辑
	for {
		 // 不要在饥饿模式下自旋，因为所有权会交给等待者，我们无论如何也获取不到锁
        //old&(mutexLocked|mutexStarving) 掩码操作，检查old中是否同时包含这两个标志位

		if old&(mutexLocked|mutexStarving) == mutexLocked && runtime_canSpin(iter) {
	    // 尝试设置 mutexWoken 标志来通知 Unlock 不要唤醒其他阻塞的goroutine
			if !awoke && old&mutexWoken == 0 && old>>mutexWaiterShift != 0 &&
				atomic.CompareAndSwapInt32(&m.state, old, old|mutexWoken) {
				awoke = true
			}
            runtime_doSpin()    // 执行自旋  自旋条件：锁被持有(mutexLocked)且不在饥饿模式(mutexStarving)
            iter++             // 增加自旋迭代计数
            old = m.state      // 重新获取状态
            continue           // 继续循环
			continue
		}
		new := old
	// 不要尝试获取处于饥饿模式的锁，新来的goroutine必须排队
		if old&mutexStarving == 0 {
			new |= mutexLocked   // 如果不是饥饿模式，尝试获取锁
		}
           // 如果锁已被持有或处于饥饿模式，增加等待者计数
		if old&(mutexLocked|mutexStarving) != 0 {
			new += 1 << mutexWaiterShift
		}
    // 当前goroutine将锁切换到饥饿模式
    // 但如果锁当前是解锁状态，不要切换
    // Unlock期望饥饿模式的锁有等待者，而这种情况不会成立
		if starving && old&mutexLocked != 0 {
			new |= mutexStarving
		}
		if awoke {
            // goroutine已从睡眠中被唤醒，所以无论如何都需要重置标志
			// 唤醒标志(mutexWoken)的生命周期：
            // 1. 设置时机：在自旋或等待时，goroutine告诉解锁者"我已经准备好了"
            // 2. 目的：优化唤醒，避免解锁时唤醒多个等待者
            // 3. 重置时机：goroutine被唤醒后，必须立即清除这个标志

            // 如果不重置会怎样？
            // 假设：goroutine A被唤醒，获取了锁，但没有清除mutexWoken标志
            // 然后goroutine A解锁时看到mutexWoken=1，以为有goroutine已准备好
            // 结果：解锁者不会唤醒其他等待者，导致死锁！
			if new&mutexWoken == 0 {
				throw("sync: inconsistent mutex state")
			}
			new &^= mutexWoken
		}
		if atomic.CompareAndSwapInt32(&m.state, old, new) {
            //如果旧状态既没有锁定也没有饥饿，说明我们成功获取了锁
			if old&(mutexLocked|mutexStarving) == 0 {
				break  // 通过CAS成功锁定了互斥锁
			}
			// 如果我们之前已经在等待，就排到队列前面（LIFO）
			queueLifo := waitStartTime != 0 // 之前等待过 → 队列前面
			if waitStartTime == 0 {
				waitStartTime = runtime_nanotime()// 记录开始等待时间
			}
            // 在信号量上等待（可能阻塞）
			runtime_SemacquireMutex(&m.sema, queueLifo, 2)
            // 更新饥饿状态：如果已经饥饿或等待时间超过阈值1ms
			starving = starving || runtime_nanotime()-waitStartTime > starvationThresholdNs
			old = m.state // 重新获取状态
			if old&mutexStarving != 0 {
            // 如果这个goroutine被唤醒且互斥锁处于饥饿模式，
            // 所有权已经移交给我们，但互斥锁处于某种不一致的状态：
            // mutexLocked 没有设置，我们仍然被计为等待者。修复这个问题。
				if old&(mutexLocked|mutexWoken) != 0 || 

                // 将 old 右移3位，提取等待者计数部分
                old>>mutexWaiterShift == 0 { 
					throw("sync: inconsistent mutex state")
				}
                 // 计算状态变化量：设置锁定标志，减少等待者计数
				delta := int32(mutexLocked - 1<<mutexWaiterShift)
                  // 如果不饥饿或者只有一个等待者，退出饥饿模式
				if !starving || old>>mutexWaiterShift == 1 {
					delta -= mutexStarving
				}
				atomic.AddInt32(&m.state, delta)
				break   // 成功获取锁
			}
            // 被唤醒后重置状态
			awoke = true
			iter = 0
		} else {
            // CAS失败，重新获取状态并继续尝试
			old = m.state
		}
	}

	if race.Enabled {
		race.Acquire(unsafe.Pointer(m))
	}
}

// 饥饿模式触发条件：
// 1. goroutine等待时间 > 1ms (starvationThresholdNs)
// 2. 当前锁被持有

// 饥饿模式特点：
// 1. 新来的goroutine直接排队，不自旋
// 2. 解锁时锁直接交给等待队列头部的goroutine
// 3. 防止新goroutine"插队"，保证公平性

// 退出饥饿模式条件：
// 1. 当前goroutine不饥饿（等待时间 < 1ms）
// 2. 它是队列中最后一个等待者

```


 互斥锁有两种操作模式：正常模式和饥饿模式。
 在正常模式下，等待者按先进先出 (FIFO) 的顺序排队，但被唤醒的等待者
 并不拥有互斥锁，而是与新到达的 goroutine 竞争
 互斥锁的所有权。新到达的 goroutine 具有优势——它们
 已经在 CPU 上运行，而且数量可能很多，因此被唤醒的等待者
 很有可能输掉竞争。在这种情况下，它会被排在等待队列的前端
 如果等待者超过 1 毫秒未能获取互斥锁，
 则互斥锁切换到饥饿模式。

 在饥饿模式下，互斥锁的所有权直接从
 解锁互斥锁的 goroutine 传递给队列前端的等待者。
 即使互斥锁看起来已解锁，新到达的 goroutine 也不会尝试获取它，也不会尝试自旋。相反，它们会把自己排在
 等待队列的尾部。
 如果一个等待者获得了互斥锁的所有权，并且发现
 (1) 它是队列中的最后一个等待者，或者 (2) 它等待的时间少于 1 毫秒，
 则它将互斥锁切换回正常操作模式。

 正常模式的性能要好得多，因为即使有等待者阻塞，goroutine 也可以连续多次获取互斥锁。
 饥饿模式对于防止出现极端的尾部延迟情况至关重要。


Unlock 操作详解
其中`-mutexLocked ` 是原子减法操作用于清除互斥锁的锁定标志
atomic.AddInt32(&m.state, -mutexLocked)
等价于：
atomic.AddInt32(&m.state, -1)
也就是：m.state = m.state - 1
`CAS 操作`： CAS = Compare And Swap（比较并交换）原子操作：如果当前值等于预期值，则更新为新值，否则不更新

``` go
func (m *Mutex) Unlock() {
    // 竞态检测支持（仅在 -race 模式下启用）
    if race.Enabled {
        _ = m.state  // 确保 m 不被编译器优化掉
        race.Release(unsafe.Pointer(m))
    }

    // 快速路径：直接清除锁定标志位
    // 使用原子减法：state = state - mutexLocked (即 state - 1)
    new := atomic.AddInt32(&m.state, -mutexLocked)
    
    // 如果 new != 0，说明还有其他状态需要处理（有等待者等）
    if new != 0 {
        // 慢速路径单独作为一个函数，便于快速路径内联优化
        m.unlockSlow(new)
    }
}

func (m *Mutex) unlockSlow(new int32) {
    // 解锁再加锁，结果为锁定，以此来判断几解锁前是否有锁定标志
	if (new+mutexLocked)&mutexLocked == 0 {
		fatal("sync: unlock of unlocked mutex")
	}
    //非饥饿模式处理
	if new&mutexStarving == 0 {
		old := new
		for {
             // 情况1：没有等待者 (old>>mutexWaiterShift == 0)
            // 情况2：锁已被其他goroutine获取 (old&mutexLocked != 0)
            // 情况3：已有goroutine被唤醒 (old&mutexWoken != 0)
            // 情况4：进入了饥饿模式 (old&mutexStarving != 0)
            // 以上任何情况都不需要唤醒其他goroutine
			if old>>mutexWaiterShift == 0 || old&(mutexLocked|mutexWoken|mutexStarving) != 0 {
				return
			}

            // 尝试获取唤醒其他goroutine的权利：
            // 1. 减少一个等待者计数 (old - 1<<mutexWaiterShift)
            // 2. 设置唤醒标志 (| mutexWoken)
            // 这样做的目的是告诉其他潜在的解锁者："我已经在唤醒一个goroutine了"
			new = (old - 1<<mutexWaiterShift) | mutexWoken
             // 原子地更新状态
			if atomic.CompareAndSwapInt32(&m.state, old, new) {
                // 唤醒一个等待的goroutine
                // 参数说明：
                // &m.sema: 信号量
                // false: 不跳过handoff（正常模式）
                // 2: 跟踪跳过帧数（用于调试）
				runtime_Semrelease(&m.sema, false, 2)
				return
			}
            // CAS失败，说明状态在更新期间被其他goroutine修改了
            // 重新获取状态并重试
			old = m.state
		}
	} else {
        // 饥饿模式：直接将锁的所有权交给下一个等待者
        // 参数说明：
        // &m.sema: 信号量
        // true: 跳过handoff（饥饿模式特殊处理）
        // 2: 跟踪跳过帧数
		runtime_Semrelease(&m.sema, true, 2)
	}
}


```
#### Mutex 拷贝特性
Go函数值传递的特点，sync.Mutex通过函数传递时，会进行一次拷贝，所以传递过去的锁是一把全新的锁

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== 验证1: 复制锁导致失效 ===")
	testCopyLock()

	fmt.Println("\n=== 验证2: 指针传递锁 ===")
	testPointerLock()

	fmt.Println("\n=== 验证3: 嵌入结构体中的复制 ===")
	testEmbeddedCopy()
}

// 验证1: 复制锁导致失效
func testCopyLock() {
	var mu sync.Mutex
	var counter int
	// 启动5个goroutine尝试获取锁
	for i := 0; i < 5; i++ {
		go func(id int, m sync.Mutex) { // 这里复制了锁！
			m.Lock()
			defer m.Unlock()

			// 这里操作的 counter 是外部的，但锁是复制的
			// 每个goroutine都有自己的锁，所以无法互斥
			fmt.Printf("Goroutine %d 获取到锁，counter=%d\n", id, counter)
			counter++
			time.Sleep(5 * time.Second)
		}(i, mu) // 将 mu 复制传递
	}

	time.Sleep(2 * time.Second)
	fmt.Printf("最终 counter=%d (应该为1，但实际上每个goroutine都能获取到'锁')\n", counter)
}

// 验证2: 使用指针传递锁
func testPointerLock() {
	var mu sync.Mutex
	var counter int

	// 启动5个goroutine，使用指针传递
	for i := 0; i < 5; i++ {
		go func(id int, m *sync.Mutex) { // 传递指针，不复制
			m.Lock()
			defer m.Unlock()

			fmt.Printf("Goroutine %d 获取到锁，counter=%d\n", id, counter)
			counter++
			time.Sleep(5 * time.Second)
		}(i, &mu) // 传递锁的地址
	}

	time.Sleep(2 * time.Second)
	fmt.Printf("最终 counter=%d (正确，只有一个goroutine能获取到锁)\n", counter)
}

// 验证3: 嵌入结构体中的复制
func testEmbeddedCopy() {
	type Container struct {
		sync.Mutex // 嵌入 Mutex
		value      int
	}

	container := Container{value: 0}

	// 启动goroutine
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int, c Container) { // 这里复制了整个Container，包括Mutex！
			c.Lock() // 使用的是复制的锁
			defer c.Unlock()

			fmt.Printf("Goroutine %d 操作 container，value=%d\n", id, c.value)
			c.value++ // 修改的是复制的container，不是原来的！

			done <- true
			time.Sleep(5 * time.Second)
		}(i, container) // 复制container
	}

	// 等待所有goroutine完成
	for i := 0; i < 5; i++ {
		<-done
	}

	fmt.Printf("最终 container.value=%d (应该为0，因为每个goroutine修改的都是自己的副本)\n", container.value)
}

```

### RWMutex 







