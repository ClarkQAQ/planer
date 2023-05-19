package planer

import (
	"sort"
	"sync"
	"time"
)

type Jobs struct {
	sort bool
	jobs []*Job
	lock *sync.Mutex
}

func (l Jobs) Len() int {
	return len(l.jobs)
}

func (l Jobs) Less(i, j int) bool {
	return l.jobs[i].Unix < l.jobs[j].Unix
}

func (l *Jobs) Swap(i, j int) {
	l.jobs[i], l.jobs[j] = l.jobs[j], l.jobs[i]
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

	j.sort = false
	j.jobs = append(j.jobs, jb)
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

	if !j.sort {
		sort.Sort(j)
		j.sort = true
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
