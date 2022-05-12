package awsu

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type Pool interface {
	Submit(j func())
	RunAll()
}

type WorkerPool struct {
	jobs       []func()
	workersNum int
	inQueue    chan func()
	stopWorker chan bool
}

func NewWorkerPool(workers int) Pool {
	return &WorkerPool{
		workersNum: workers,
		jobs:       []func(){},
		inQueue:    make(chan func()),
		stopWorker: make(chan bool),
	}
}

func (wp *WorkerPool) Submit(j func()) {
	wp.jobs = append(wp.jobs, j)
}

func (wp *WorkerPool) runSingleWorker(wg *sync.WaitGroup) {
	log.Debug("started single worker")
	for {
		select {
		case j, ok := <-wp.inQueue:
			if !ok {
				return
			}
			j()
			log.Debug("job done")
			wg.Done()
		case <-wp.stopWorker:
			return
		}
	}

}

func (wp *WorkerPool) RunAll() {

	var wg sync.WaitGroup
	wg.Add(len(wp.jobs))
	for i := 0; i < wp.workersNum; i++ {
		go wp.runSingleWorker(&wg)
	}

	for i := 0; i < len(wp.jobs); i++ {
		wp.inQueue <- wp.jobs[i]
	}

	log.Debug("waiting for workes to finish")

	wg.Wait()

	wp.stopWorker <- true
	close(wp.inQueue)
	close(wp.stopWorker)

}
