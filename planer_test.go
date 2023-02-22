package planer

import (
	"math/rand"
	"sort"
	"sync"
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

	select {
	case <-ch:
	case <-time.After(time.Second * 2):
		t.Error("timeout")
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

func TestJobs_insert(t *testing.T) {
	j := &Jobs{
		jobs: []*Job{},
		lock: &sync.Mutex{},
	}

	data := []int64{
		5, 2, 3, 4, 1, 10, 9, 8, 7, 6,
		12, 11, 15, 14, 13, 16, 17, 18, 19, 20,
		30, 29, 28, 27, 26, 21, 22, 23, 24, 25,
	}

	for _, v := range data {
		j.Insert(&Job{
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

func Benchmark_insert(b *testing.B) {
	j := &Jobs{
		jobs: []*Job{},
		lock: &sync.Mutex{},
	}

	mm := []int64{}
	for i := 0; i < 100; i++ {
		mm = append(mm, rand.Int63n(int64(b.N)))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j.Insert(&Job{
			Unix: mm[i%100],
		})
	}
}
