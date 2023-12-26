package planer

import (
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestPlaner_New(t *testing.T) {
	New().AddJob(0, func() {})
}

func TestPlaner_Start(t *testing.T) {
	p := New()
	p.SetWaitDuration(time.Second)

	ch := make(chan bool)
	p.AddJob(time.Now().Unix()+1, func() {
		t.Log("hello")
		ch <- true
	})

	p.Start()

	// 多次启动
	p.Start()

	select {
	case <-ch:
	case <-time.After(time.Second * 2):
		t.Error("timeout")
	}
}

// [Bug] 任务队列顺序问题 #1
func TestPlaner_BugJob1(t *testing.T) {
	p := New()
	p.SetWaitDuration(time.Second)

	ch1, chd10 := make(chan bool), make(chan bool)
	chu300 := make(chan bool)
	p.AddJob(time.Now().Unix()+1, func() {
		ch1 <- true

		p.AddJob(time.Now().Unix()-10, func() {
			chd10 <- true
		})

		p.AddJob(time.Now().Unix()+300, func() {
			chu300 <- true
		})
	})

	p.Start()
	defer p.Stop()

	time.Sleep(time.Second * 2)

	ch2 := make(chan bool)
	p.AddJob(time.Now().Unix()+1, func() {
		ch2 <- true
	})

	select {
	case <-ch1:
	case <-time.After(time.Second):
		t.Error("ch1 timeout")
	}

	select {
	case <-chd10:
	case <-time.After(time.Second):
		t.Error("chd10 timeout")
	}

	select {
	case <-ch2:
	case <-time.After(time.Second * 2):
		t.Error("ch2 timeout")
	}

	select {
	case <-chu300:
		t.Error("chu300 should not be here")
	default:
	}
}

func TestPlaner_Stop(t *testing.T) {
	p := New()
	p.SetWaitDuration(time.Second)

	ch := make(chan bool, 1)
	p.AddJob(time.Now().Unix()+2, func() {
		t.Log("hello")
		ch <- true
	})

	p.Start()
	time.Sleep(time.Second)
	p.Stop()

	select {
	case <-ch:
		t.Error("should not be here")
	case <-time.After(time.Second * 2):
	}
}

func TestPlaner_StopAfter(t *testing.T) {
	p := New()
	p.SetWaitDuration(time.Second)

	ch := make(chan bool, 1)
	p.AddJob(time.Now().Unix()+3, func() {
		t.Log("hello")
		ch <- true
	})

	p.Start()
	time.Sleep(time.Second)
	p.Stop()
	p.Start()

	select {
	case <-ch:
		t.Error("should not be here")
	case <-time.After(time.Second * 2):
	}
}

func BenchmarkPlaner_AddJob(b *testing.B) {
	p := New()
	cb := func() {}
	unix := time.Now().Unix()

	b.ResetTimer()

	n := int64(b.N)
	for i := int64(0); i < n; i++ {
		p.AddJob(unix+i, cb)
	}
}

func TestJobs_Range(t *testing.T) {
	j := newJobs()

	for i := 0; i < 100000; i++ {
		j.insert(&Job{
			Unix: rand.Int63n(10000000000),
		})
	}

	j.insert(&Job{
		Unix: 1,
	})

	if j.pop().Unix != 1 {
		t.Error("pop error")
	}
}

func TestJobs_Insert(t *testing.T) {
	j := newJobs()

	data := []int64{
		5, 2, 3, 4, 1, 10, 9, 8, 7, 6,
		12, 11, 15, 14, 13, 16, 17, 18, 19, 20,
		30, 29, 28, 27, 26, 21, 22, 23, 24, 25,
	}

	for _, v := range data {
		j.insert(&Job{
			Unix: v,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i] < data[j]
	})

	for i := 0; i < len(data); i++ {
		if data[i] != j.pop().Unix {
			t.Errorf("jobs[%d] = %d", i, j.jobs[i].Unix)
		}
	}
}

func BenchmarkJobs_Insert(b *testing.B) {
	j := newJobs()

	mm := []int64{}
	for i := 0; i < 100; i++ {
		mm = append(mm, rand.Int63n(int64(10000000000)))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j.insert(&Job{
			Unix: mm[i%100],
		})
	}
}

func Benchmark_Pop(b *testing.B) {
	j := newJobs()

	for i := 0; i < 100000; i++ {
		j.insert(&Job{
			Unix: rand.Int63n(10000000000),
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j.pop()
	}
}

// [Performance] 在短时间插入大量无序任务后执行耗时很久 #2
func Benchmark_PopWithInsert(b *testing.B) {
	j := newJobs()

	mm := []int64{}
	for i := 0; i < 100; i++ {
		mm = append(mm, rand.Int63n(int64(10000000000)))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j.insert(&Job{
			Unix: mm[i%100],
		})

		if i%100 == 0 {
			j.pop()
		}
	}
}
