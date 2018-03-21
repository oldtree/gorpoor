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

func Exec() error {
	log.Println("music")
	return nil
}

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
		w.Stop()
		//log.Printf("worker [%d] is end statu [%d]\n", w.WorkerId, w.Status)
	}()
	if atomic.LoadInt64(&w.Status) == STATUS_START {
		atomic.StoreInt64(&w.Status, STATUS_RUNNING)
	}
	var err error
	for {
		select {
		case <-w.StopChan:
			goto END
		case t, ok := <-w.TaskList:

			if ok && t != nil {
				err = t.Exec()
				if err != nil {
					log.Println("exec task error", err.Error())
				}
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

	Protect sync.Mutex
}

func (w *WorkerPool) Init(number int, taskLength int) {
	w.TaskList = make(chan Tasker, taskLength)
	w.StopChan = make(chan struct{}, 1)
	atomic.StoreInt64(&w.Status, STATUS_INIT)
	w.WorkQueue = make([]*Worker, number, number)
	w.Wg = new(sync.WaitGroup)
	for index, _ := range w.WorkQueue {
		newWorker := new(Worker)
		newWorker.Wg = w.Wg
		newWorker.StopChan = w.StopChan
		newWorker.TaskList = w.TaskList
		newWorker.Init(int64(index))
		w.WorkQueue[index] = newWorker
		go w.WorkQueue[index].Start()
	}

	atomic.StoreInt64(&w.Status, STATUS_INIT)
	return
}
func (w *WorkerPool) Start() {
	defer func() {
	}()
	for {
		select {
		case <-w.StopChan:
			goto END
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
	w.StopChan <- struct{}{}
	atomic.StoreInt64(&w.Status, STATUS_STOP)
	close(w.StopChan)
	close(w.TaskList)
	w.Wg.Wait()
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
