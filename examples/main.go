package main

import (
	"fmt"
	"time"

	"github.com/ClarkQAQ/planer"
)

func main() {
	p := planer.New()
	p.SetWaitDuration(time.Second) // 设置没有任务时的循环时间

	t := time.Now()

	// 添加任务
	p.AddJob(time.Now().Unix()+1, func() {
		p.AddJob(time.Now().Unix()+2, func() {
			fmt.Printf("hello world 1+2, since: %s\n", time.Since(t))
		})
		fmt.Printf("hello world 1, since: %s\n", time.Since(t))
	})

	p.AddJob(time.Now().Unix()+300, func() {
		fmt.Println("hello world 300")
	})

	// 启动
	p.Start()
	defer p.Stop() // 停止, 如果有没有执行的任务将会清空

	time.Sleep(time.Second * 2)

	// 添加任务
	p.AddJob(time.Now().Unix()+5, func() {
		p.AddJob(time.Now().Unix()+2, func() {
			fmt.Printf("hello world 5+2+2, since: %s\n", time.Since(t))
		})
		fmt.Printf("hello world 5+2, since: %s\n", time.Since(t))
	})

	// 等待
	time.Sleep(time.Second * 10)
}
