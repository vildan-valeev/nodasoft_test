package pool

import (
	"fmt"
	"sync"
)

// Pool is the worker pool
type Pool struct {
	Workers []*Worker

	concurrency   int
	collector     chan *Task
	runBackground chan bool
	wg            sync.WaitGroup
}

// NewPool initializes a new pool with the given tasks and
// at the given concurrency.
func NewPool(concurrency int) *Pool {
	return &Pool{
		concurrency: concurrency,
		collector:   make(chan *Task, 1000),
	}
}

// AddTask adds a task to the pool
func (p *Pool) AddTask(task *Task) {
	p.collector <- task
}

// Run runs all work within the pool and blocks until it's
// finished.
func (p *Pool) Run(s string) {
	fmt.Printf("Запуск пула - %s \n", s)
	for i := 1; i <= p.concurrency; i++ {
		worker := NewWorker(p.collector, i)
		p.Workers = append(p.Workers, worker)
		go worker.Start()
	}

	p.runBackground = make(chan bool)
	<-p.runBackground
}

// Stop stops background workers
func (p *Pool) Stop() {
	for i := range p.Workers {
		p.Workers[i].Stop()
	}
	p.runBackground <- true
}
