package v1

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
	ErrWorkPoolIsRuning = errors.New("work pool is running")
)

type Tasker interface {
	Exec() error
}

type Task struct {
	Params interface{}
	Result interface{}
}

func (t *Task) Exec() error {
	log.Println("task exec result ok")
	return nil
}

type Worker struct {
	WorkerId int64
	StopChan chan struct{}
	TaskList chan Tasker

	Status int64
	Wg     *sync.WaitGroup
}

func (w *Worker) Init(index int64) {
	atomic.StoreInt64(&w.Status, STATUS_INIT)
	w.WorkerId = index
	w.Wg.Add(1)
	atomic.StoreInt64(&w.Status, STATUS_START)
}

func (w *Worker) Start() {
	defer func() {
		log.Println("worker [%d] is end \n", w.WorkerId)
	}()
	if atomic.LoadInt64(&w.Status) == STATUS_INIT || atomic.LoadInt64(&w.Status) == STATUS_START {
		atomic.StoreInt64(&w.Status, STATUS_RUNNING)
	} else {
		return
	}
	for {
		select {
		case <-w.StopChan:
			w.Stop()
			goto END
		case t := <-w.TaskList:
			log.Println("start exec new task", t.Exec())
		}
	}
END:
	log.Printf("worker [%d] is end \n", w.WorkerId)
	return
}

func (w *Worker) Stop() {
	if atomic.LoadInt64(&w.Status) == STATUS_STOP {
		return
	}
	atomic.StoreInt64(&w.Status, STATUS_STOP)
	w.Wg.Done()
	return
}

type WorkerPool struct {
	TaskList chan Tasker
	StopChan chan struct{}

	WorkQueue []*Worker

	Status int64
	Wg     *sync.WaitGroup
}

func (w *WorkerPool) Init(number int, taskLength int) {
	w.TaskList = make(chan Tasker, taskLength)
	w.StopChan = make(chan struct{}, 1)
	atomic.StoreInt64(&w.Status, STATUS_INIT)
	w.WorkQueue = make([]*Worker, number, number)
	w.Wg = new(sync.WaitGroup)
	for index, _ := range w.WorkQueue {
		w.WorkQueue[index] = new(Worker)
		w.WorkQueue[index].Wg = w.Wg
		w.WorkQueue[index].StopChan = w.StopChan
		w.WorkQueue[index].TaskList = w.TaskList
		w.WorkQueue[index].Init(int64(index))
		go w.WorkQueue[index].Start()
	}
	w.Wg.Add(number)
	atomic.StoreInt64(&w.Status, STATUS_INIT)
	return
}
func (w *WorkerPool) Start() {
	defer func() {
		log.Println("work pool is end")
		w.Stop()
	}()
	for {
		select {
		case <-w.StopChan:
			w.Stop()
			goto END
		}
	}
END:
	log.Println("work pool is stop")
	return
}
func (w *WorkerPool) Stop() {
	if atomic.LoadInt64(&w.Status) == STATUS_STOP {
		return
	}
	w.StopChan <- struct{}{}
	close(w.StopChan)
	close(w.TaskList)
	atomic.StoreInt64(&w.Status, STATUS_STOP)
	w.Wg.Wait()
	return
}

func (w *WorkerPool) AddTask(t Tasker) {
	if atomic.LoadInt64(&w.Status) == STATUS_STOP {
		log.Println("work pool is stop")
		return
	}
	w.TaskList <- t
	return
}
