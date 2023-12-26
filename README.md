<p align="center">
  <h3 align="center">Planer</h3>
  <p align="center">
    内部业务使用到的轻量级时间戳计划任务
  </p>
</p>


## 1. 介绍

Planer 是一个轻量级基于时间戳的精确计划任务, 可以用于业务的计划任务, 比如 `2023-02-22 11:33:12` 执行某个任务.

为什么不用 `cron` 呢？因为普通的年龄 40+ 人类难以理解  `cron`, 哪怕做好 UI 提示, 还是多个选项让他们选择目标时间比较简单.

## 2. 使用

超级简单, 最多只需要调五个函数就行...

```go

package main

import (
	"fmt"
	"time"

	"github.com/ClarkQAQ/planer"
)

func main() {
	p := planer.New()
	p.SetWaitDuration(time.Second) // 设置没有任务时的循环时间

	// 添加任务
	p.AddJob(time.Now().Unix()+3, func() {
		fmt.Println("hello world")
	})

	// 启动
	p.Start()
	defer p.Stop() // 停止, 如果有没有执行的任务将会清空

	// 等待
	time.Sleep(time.Second * 5)
}

```

3. Benchmark

删除了插入排序改成了读取排序, 插入速度提升了, 读取速度虽慢, 但是对于一个秒级的计划任务来说, 并没影响

```go

[clark@ArchOwO planer]$ go test -benchmem -run=^$ -bench ^Benchmark.*$ github.com/ClarkQAQ/planer
goos: linux
goarch: amd64
pkg: github.com/ClarkQAQ/planer
cpu: AMD Ryzen 7 5800H with Radeon Graphics         
BenchmarkPlaner_AddJob-16       17171708                78.37 ns/op           60 B/op          1 allocs/op
BenchmarkJobs_Insert-16         14351600                85.20 ns/op           58 B/op          1 allocs/op
Benchmark_Pop-16                206814886                5.516 ns/op           0 B/op          0 allocs/op
Benchmark_PopWithInsert-16      11885229               111.2 ns/op            57 B/op          1 allocs/op
PASS
ok      github.com/ClarkQAQ/planer      5.671s

```