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
	atomic.StoreInt64(&w.Status, STATUS_INIT)
	w.Wg.Add(1)
}

func (w *Worker) Start() {
	defer func() {
		log.Printf("worker [%d] is end \n", w.WorkerId)
		w.Stop()
	}()
	var err error
	for {
		select {
		case <-w.StopChan:
			w.Stop()
			return
		case t := <-w.TaskList:
			err = t.Exec()
			if err != nil {
				log.Println("exec task failed : ", err.Error())
			}
		}
	}
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
		w.WorkSlice[index] = new(Worker)
		w.WorkSlice[index].Wg = w.Wg
		w.WorkSlice[index].Init(index, taskLength)
		go w.WorkSlice[index].Start()
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
		log.Println("work pool is end")
	}()
	for {
		select {
		case <-w.StopChan:
			goto END_POOL
		case t := <-w.TaskList:
			index := randomWorld.Int31n(int32(w.WorkNumber))
			w.WorkSlice[int(index)].TaskList <- t
		}
	}
END_POOL:
	log.Println("work pool is end")
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
