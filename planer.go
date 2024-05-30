package planer

import (
	"container/heap"
	"sync"
	"time"
)

type Jobs struct {
	jobs []*Job
	lock *sync.Mutex
}

func (j *Jobs) Len() int {
	return len(j.jobs)
}

func (j *Jobs) Less(i, k int) bool {
	return j.jobs[i].Unix < j.jobs[k].Unix
}

func (j *Jobs) Swap(i, k int) {
	j.jobs[i], j.jobs[k] = j.jobs[k], j.jobs[i]
}

func (j *Jobs) Push(x interface{}) {
	j.jobs = append(j.jobs, x.(*Job))
}

func (j *Jobs) Pop() interface{} {
	n := len(j.jobs)
	x := j.jobs[n-1]
	j.jobs = j.jobs[:n-1]
	return x
}

type Job struct {
	Unix int64
	Job  func()
}

func newJobs() *Jobs {
	j := &Jobs{
		jobs: make([]*Job, 0, 512),
		lock: &sync.Mutex{},
	}

	j.clean()

	return j
}

func (j *Jobs) insert(jb *Job) {
	j.lock.Lock()
	defer j.lock.Unlock()
	heap.Push(j, jb)
}

func (j *Jobs) clean() {
	j.lock.Lock()
	j.jobs = j.jobs[:0]
	heap.Init(j)
	j.lock.Unlock()
}

func (j *Jobs) pop() *Job {
	j.lock.Lock()
	defer j.lock.Unlock()

	if len(j.jobs) < 1 {
		return nil
	}

	return heap.Pop(j).(*Job)
}

type Planer struct {
	j            *Jobs
	timer        *time.Timer
	currentLock  *sync.Mutex
	currentJob   *Job
	signal       chan bool
	waitDuration time.Duration
}

func New() *Planer {
	return &Planer{
		j:            newJobs(),
		timer:        nil,
		currentLock:  &sync.Mutex{},
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
	p.j.insert(&Job{
		Unix: unix,
		Job:  job,
	})

	if p.timer == nil {
		return
	}

	p.currentLock.Lock()
	defer p.currentLock.Unlock()

	if p.currentJob != nil {
		if unix > p.currentJob.Unix {
			return
		}

		p.j.insert(p.currentJob)
	}

	p.currentJob = p.j.pop()
	p.timer.Reset(0)
}

func (p *Planer) Start() {
	p.currentLock.Lock()

	if p.timer != nil {
		p.currentLock.Unlock()
		return
	}

	go func(fn func(), unlock func()) {
		defer fn()
		unlock()
	}(p.run, p.currentLock.Unlock)
}

func (p *Planer) ticker(now time.Time) {
	p.currentLock.Lock()
	defer p.currentLock.Unlock()

	if p.currentJob != nil {
		if p.currentJob.Unix-now.Unix() > 0 {
			p.timer.Reset(time.Duration(p.currentJob.Unix-now.Unix()) * time.Second)
			return
		}

		go func(fn func()) { fn() }(p.currentJob.Job)
	}

	if p.currentJob = p.j.pop(); p.currentJob != nil {
		p.timer.Reset(0)
		return
	}

	p.timer.Reset(p.waitDuration)
}

func (p *Planer) run() {
	p.timer = time.NewTimer(0)
	for {
		select {
		case <-p.signal:
			p.timer.Stop()
			p.timer = nil
			return
		case now := <-p.timer.C:
			p.ticker(now)
		}
	}
}

func (p *Planer) Stop() {
	if p.timer != nil {
		p.signal <- true
	}

	p.j.clean()
}
