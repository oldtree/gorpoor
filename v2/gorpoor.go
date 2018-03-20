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
	atomic.StoreInt64(&w.WorkerId, STATUS_INIT)
}

func (w *Worker) Start() {
	defer func() {
		log.Printf("work [%d] loop is end \n", w.WorkerId)
		return
	}()
	atomic.StoreInt64(&w.Status, STATUS_RUNNING)
	for {
		select {
		case <-w.StopChan:
			w.Stop()
			goto END
		case t := <-w.TaskChan:
			log.Println("task exec ", t.Exec())
			w.WorkerBackChan <- w
		}
	}
END:
	log.Printf("work [%d] loop is end \n", w.WorkerId)
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

type WorkorPool struct {
	TaskList    chan Tasker
	Status      int64
	Worker      chan *Worker
	WorkorQueue []*Worker
	Wg          *sync.WaitGroup

	StopChan chan struct{}
}

func (w *WorkorPool) Init(number int, taskLength int) {
	w.Worker = make(chan *Worker, number)
	w.Wg = new(sync.WaitGroup)
	w.WorkorQueue = make([]*Worker, number)
	for index, _ := range w.WorkorQueue {
		w.WorkorQueue[index] = new(Worker)
		w.WorkorQueue[index].Init(index)
		go w.WorkorQueue[index].Start()
	}
	w.TaskList = make(chan Tasker, taskLength)
	atomic.StoreInt64(&w.Status, STATUS_INIT)
	w.StopChan = make(chan struct{}, 1)
	return
}

func (w *WorkorPool) Start() {
	defer func() {
		log.Println("work pool is end")
	}()
	atomic.StoreInt64(&w.Status, STATUS_START)
	for {
		select {
		case <-w.StopChan:
			goto END
		case t := <-w.TaskList:
			worker := <-w.Worker
			worker.TaskChan <- t
		}
	}
END:
	log.Println("work pool is end")
}

func (w *WorkorPool) Stop() {
	if atomic.LoadInt64(&w.Status) == STATUS_STOP {
		return
	}
	atomic.StoreInt64(&w.Status, STATUS_STOP)
	close(w.StopChan)
	close(w.TaskList)
	close(w.Worker)
	w.Wg.Wait()
	return
}

func (w *WorkorPool) AddTask(t Tasker) {
	if atomic.LoadInt64(&w.Status) != STATUS_RUNNING {
		return
	}
	w.TaskList <- t
	return
}
