package v2

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
)

const (
	STATUS_INIT = iota
	STATUS_START
	STATUS_RUNNING
	STATUS_STOP
)

var (
	ErrorTaskExec = errors.New("task exec error")
)

type Execer func() error

func (exec Execer) Exec() error {
	return exec()
}

type Tasker interface {
	Exec() error
}

type Task struct {
	Params interface{}
	Result interface{}
}

type Worker struct {
	WorkerId int64
	StopChan chan struct{}

	TaskChan       chan Tasker
	WorkerBackChan chan *Worker

	Wg     *sync.WaitGroup
	Status int64
}

func (w *Worker) Init(index int) {
	w.WorkerId = int64(index)
	w.TaskChan = make(chan Tasker, 1)
	w.WorkerBackChan <- w
	w.Wg.Add(1)
	atomic.StoreInt64(&w.WorkerId, STATUS_INIT)
}

func (w *Worker) Start() {
	defer func() {
		w.Stop()
		return
	}()
	atomic.StoreInt64(&w.Status, STATUS_RUNNING)
	var err error
	for {
		select {
		case <-w.StopChan:
			goto END
		case t, ok := <-w.TaskChan:
			if ok && t != nil {
				err = t.Exec()
				if err != nil {
					log.Println("task exec error : ", err.Error())
				}
				w.WorkerBackChan <- w
			} else {
				goto END
			}

		}
	}
END:
	return
}

func (w *Worker) Stop() {
	if atomic.LoadInt64(&w.Status) == STATUS_STOP {
		return
	}
	close(w.TaskChan)
	atomic.StoreInt64(&w.Status, STATUS_STOP)
	w.Wg.Done()
	return
}

type WorkerPool struct {
	TaskList    chan Tasker
	Status      int64
	Worker      chan *Worker
	WorkerQueue []*Worker
	Wg          *sync.WaitGroup
	Protect     sync.Mutex
	StopChan    chan struct{}
}

func (w *WorkerPool) Init(number int, taskLength int) {
	w.Worker = make(chan *Worker, number)
	w.Wg = new(sync.WaitGroup)
	w.WorkerQueue = make([]*Worker, number)
	w.StopChan = make(chan struct{}, 1)
	w.TaskList = make(chan Tasker, taskLength)
	for index, _ := range w.WorkerQueue {
		newWorker := new(Worker)
		newWorker.Wg = w.Wg
		newWorker.WorkerBackChan = w.Worker
		newWorker.StopChan = w.StopChan
		newWorker.Init(index)
		w.WorkerQueue[index] = newWorker
		go newWorker.Start()
	}
	atomic.StoreInt64(&w.Status, STATUS_INIT)
	return
}

func (w *WorkerPool) Start() {
	defer func() {
		//log.Println("work pool is end")
	}()
	atomic.StoreInt64(&w.Status, STATUS_START)
	for {
		select {
		case <-w.StopChan:
			goto END
		case t := <-w.TaskList:
			worker, ok := <-w.Worker
			if ok && worker != nil {
				worker.TaskChan <- t
			} else {
				goto END
			}
		}
	}
END:
	return
}

func (w *WorkerPool) Stop() {
	w.Protect.Lock()
	defer w.Protect.Unlock()
	if atomic.LoadInt64(&w.Status) == STATUS_STOP {
		return
	}
	atomic.StoreInt64(&w.Status, STATUS_STOP)
	close(w.StopChan)
	w.Wg.Wait()
	close(w.TaskList)
	close(w.Worker)
	return
}

func (w *WorkerPool) AddTask(t Tasker) {
	w.Protect.Lock()
	defer w.Protect.Unlock()
	if atomic.LoadInt64(&w.Status) == STATUS_STOP {
		return
	}
	w.TaskList <- t
	return
}
