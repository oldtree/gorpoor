package v3

import (
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const (
	STATUS_INIT = iota
	STATUS_START
	STATUS_RUNNING
	STATUS_STOP
)

type Execer func() error

func (exec Execer) Exec() error {
	return exec()
}

var randomWorld = rand.New(rand.NewSource(time.Now().Unix()))

type Tasker interface {
	Exec() error
}

type Task struct {
	Params interface{}
	Result interface{}
}

func (t *Task) Exec() error {
	log.Printf("start exec task : %p \n", t)
	return nil
}

type Worker struct {
	WorkerId int64
	TaskList chan Tasker
	Status   int64
	StopChan chan struct{}
	Wg       *sync.WaitGroup
}

func (w *Worker) Init(index int, taskLength int) {
	w.WorkerId = int64(index)
	w.TaskList = make(chan Tasker, taskLength)
	w.Wg.Add(1)
	atomic.StoreInt64(&w.Status, STATUS_INIT)
}

func (w *Worker) Start() {
	defer func() {
		w.Stop()
	}()
	var err error
	for {
		select {
		case <-w.StopChan:
			goto END
		case t, ok := <-w.TaskList:
			if ok {
				err = t.Exec()
				if err != nil {
					log.Println("exec task failed : ", err.Error())
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
	close(w.TaskList)
	w.Wg.Done()
}

type WorkerPool struct {
	WorkNumber int
	WorkSlice  []*Worker
	TaskList   chan Tasker

	Status int64

	StopChan chan struct{}
	Wg       *sync.WaitGroup
}

func (w *WorkerPool) Init(number int, taskLength int) {
	atomic.StoreInt64(&w.Status, STATUS_STOP)
	w.WorkNumber = number
	w.WorkSlice = make([]*Worker, w.WorkNumber, w.WorkNumber)
	w.Wg = new(sync.WaitGroup)
	w.StopChan = make(chan struct{}, 1)
	w.TaskList = make(chan Tasker, taskLength)
	for index, _ := range w.WorkSlice {
		newWorker := new(Worker)
		newWorker.Wg = w.Wg
		newWorker.Init(index, taskLength)
		newWorker.StopChan = w.StopChan
		w.WorkSlice[index] = newWorker
		go newWorker.Start()
	}
	atomic.StoreInt64(&w.Status, STATUS_INIT)
	return
}

func (w *WorkerPool) AddTask(task Tasker) {
	if atomic.LoadInt64(&w.Status) != STATUS_STOP {
		w.TaskList <- task
		return
	}
	log.Panic("error operation on stop pool")
	return
}

func (w *WorkerPool) Start() {
	atomic.StoreInt64(&w.Status, STATUS_START)
	defer func() {
		if atomic.LoadInt64(&w.Status) == STATUS_STOP {
			return
		} else {
			w.Stop()
		}
	}()
	for {
		select {
		case <-w.StopChan:
			goto END
		case t, ok := <-w.TaskList:
			if ok && t != nil {
				index := randomWorld.Int31n(int32(w.WorkNumber))
				w.WorkSlice[int(index)].TaskList <- t
			} else {
				goto END
			}
		}
	}
END:
	return
}

func (w *WorkerPool) Stop() {
	if atomic.LoadInt64(&w.Status) == STATUS_STOP {
		return
	}
	atomic.StoreInt64(&w.Status, STATUS_STOP)
	w.StopChan <- struct{}{}
	close(w.StopChan)
	w.Wg.Wait()
	return
}
