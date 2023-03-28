package planer

import (
	"sync"
	"time"
)

type Jobs struct {
	jobs []*Job
	lock *sync.Mutex
}

type Job struct {
	Unix int64
	Job  func()
}

func newJobs() *Jobs {
	return &Jobs{
		jobs: []*Job{},
		lock: &sync.Mutex{},
	}
}

func (j *Jobs) insert(jb *Job) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.jobs = append(j.jobs, jb)

	// 插入排序
	// Benchmark_insert-16    	  167454	    174432 ns/op	      55 B/op	       1 allocs/op
	// for i := 1; i < len(j.jobs); i++ {
	// 	for n := i; n > 0 && j.jobs[n].Unix < j.jobs[n-1].Unix; n-- {
	// 		j.jobs[n], j.jobs[n-1] = j.jobs[n-1], j.jobs[n]
	// 	}
	// }

	// TODO: 还是太慢了
	// 二分排序
	// Benchmark_insert-16    	   21949	    133977 ns/op	      56 B/op	       1 allocs/op
	for i := 1; i < len(j.jobs); i++ {
		low, high := 0, i-1
		for low <= high {
			mid := (low + high) / 2
			if j.jobs[i].Unix < j.jobs[mid].Unix {
				high = mid - 1
			} else {
				low = mid + 1
			}
		}
		for n := i; n > low; n-- {
			j.jobs[n], j.jobs[n-1] = j.jobs[n-1], j.jobs[n]
		}
	}
}

func (j *Jobs) clean() {
	j.lock.Lock()
	j.jobs = j.jobs[:0]
	j.lock.Unlock()
}

func (j *Jobs) pop() *Job {
	j.lock.Lock()
	defer j.lock.Unlock()

	if len(j.jobs) < 1 {
		return nil
	}

	jb := j.jobs[0]
	j.jobs = j.jobs[1:]

	return jb
}

type Planer struct {
	*Jobs

	timer        *time.Timer
	currentJob   *Job
	signal       chan bool
	waitDuration time.Duration
}

func New() *Planer {
	return &Planer{
		Jobs:         newJobs(),
		timer:        nil,
		signal:       make(chan bool),
		waitDuration: time.Second,
	}
}

// SetWaitDuration 设置空闲等待时间
// SetWaitDuration sets the idle wait time
func (p *Planer) SetWaitDuration(d time.Duration) {
	p.waitDuration = d
}

func (p *Planer) AddJob(unix int64, job func()) {
	if unix < time.Now().Unix() {
		return
	}

	p.insert(&Job{
		Unix: unix,
		Job:  job,
	})

	if p.timer == nil {
		return
	}

	if p.currentJob != nil {
		if unix > p.currentJob.Unix {
			return
		}

		p.insert(p.currentJob)
	}

	p.currentJob = p.pop()
	p.timer.Reset(time.Duration(p.currentJob.Unix-unix) * time.Second)
}

func (p *Planer) Start() {
	if p.timer != nil {
		return
	}

	go p.run()
}

func (p *Planer) run() {
	p.timer = time.NewTimer(p.waitDuration)
	p.currentJob = p.pop()
	for {
		select {
		case <-p.signal:
			p.timer.Stop()
			p.timer = nil
			return

		case now := <-p.timer.C:
			if p.currentJob != nil && p.currentJob.Unix <= now.Unix() {
				go p.currentJob.Job()
				p.currentJob = p.pop()
			}

			// 定义下次执行时间
			if p.currentJob != nil && p.currentJob.Unix-now.Unix() >= 0 {
				p.timer.Reset(time.Duration(p.currentJob.Unix-now.Unix()) * time.Second)
				continue
			}

			p.currentJob = p.pop()
			p.timer.Reset(p.waitDuration)
		}
	}
}

func (p *Planer) Stop() {
	if p.timer != nil {
		p.signal <- true
	}

	p.clean()
}
