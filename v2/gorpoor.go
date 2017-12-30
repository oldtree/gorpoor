package v2

import (
	"errors"
	"log"
	"sync/atomic"
)

const (
	STATUS_START = iota
	STATUS_RUNNING
	STATUS_STOP
	STATUS_PAUSE
	STATUS_WAITING
	STATUS_RESTART
)

var (
	ErrPoolIsRunning  = errors.New("worker pool is running")
	ErrPoolInitFailed = errors.New("worker pool init failed")
	ErrPoolIsStoped   = errors.New("worker pool is stoped")

	ErrWorkerIsWorking = errors.New("worker is running")
	ErrWorkerIsStoped  = errors.New("worker is stoped")
)

type TASK func()

type Worker interface {
	Init(chan struct{}) error
	Start() error
	Stop() error
	Busy(t TASK) error
}

type WorkerFactory func() Worker
type Rest func(chan Worker) error

type FightWorker struct {
	WorkerChan chan Worker
	StopChan   chan struct{}
	QuitChan   chan struct{}
	Status     int32
	NumberTask int32
	TaskCache  chan TASK
}

func (f *FightWorker) Init(stopChan chan struct{}) error {
	f.StopChan = stopChan
	atomic.StoreInt32(&f.NumberTask, 0)
	f.QuitChan = make(chan struct{})
	f.TaskCache = make(chan TASK, 1)
	atomic.StoreInt32(&f.Status, STATUS_START)
	return nil
}

func (f *FightWorker) Start() error {
	if atomic.LoadInt32(&f.Status) == STATUS_RUNNING {
		return ErrWorkerIsWorking
	}
	atomic.StoreInt32(&f.Status, STATUS_RUNNING)
WORKLOOP:
	for {
		select {
		case newtask := <-f.TaskCache:
			newtask()
			f.WorkerChan <- f
		case <-f.QuitChan:
			break WORKLOOP
		case _, ok := <-f.StopChan:
			if ok {
				close(f.StopChan)
			}else{
				log.
			}
		}
	}
}

func (f *FightWorker) Stop() error {
	if atomic.LoadInt32(&f.Status) == STATUS_STOP {
		return ErrWorkerIsStoped
	}
	atomic.StoreInt32(&f.Status, STATUS_STOP)
	close(f.TaskCache)
	close(f.QuitChan)
	return nil
}

func (f *FightWorker) Busy(task TASK) error {
	f.TaskCache <- task
	return nil
}

type WorkerPool struct {
	Status       int32
	NumberWorker int32
	NumberTask   int32
	Tasklist     chan TASK
	WorkerList   chan Worker
	StopChan     chan struct{}
}

func (w *WorkerPool) AddTask(t TASK) error {
	if atomic.LoadInt32(&w.Status) != STATUS_RUNNING {
		return errors.New("work pool is not running")
	}
	w.Tasklist <- t
	return nil
}

func (w *WorkerPool) Init(workerNumber int, taskNumber int, gw WorkerFactory) error {

	if atomic.LoadInt32(&w.Status) == STATUS_RUNNING || atomic.LoadInt32(&w.Status) == STATUS_START {
		return errors.New("worker pool is init")
	}
	if workerNumber <= 0 {
		workerNumber = 16
	}
	if taskNumber <= 0 {
		taskNumber = 32
	}
	atomic.StoreInt32(&w.Status, STATUS_START)
	w.Tasklist = make(chan TASK, taskNumber)
	w.WorkerList = make(chan Worker, taskNumber)
	for index := 0; index < workerNumber; index++ {
		newWorker := gw()
		newWorker.Init(w.StopChan)
		go newWorker.Start()
		w.WorkerList <- newWorker
	}
	go w.Start()
	return nil
}

func (w *WorkerPool) Start() error {
	if atomic.LoadInt32(&w.Status) == STATUS_RUNNING {
		return ErrPoolIsRunning
	}
	defer func() {
		if re := recover(); re != nil {
			log.Println("worker pool is panic : ", re)
			atomic.StoreInt32(&w.Status, STATUS_STOP)
		}
	}()
WORKLOOP:
	for {
		select {
		case newT := <-w.Tasklist:
			w := <-w.WorkerList
			w.Busy(newT)
		case _, ok := <-w.StopChan:
			if ok { //还没有关闭
				log.Println("work pool will be closed")
				close(w.StopChan)
			} else { //关闭
				log.Println("work pool is already close")
			}
			break WORKLOOP
		}
	}

	return nil
}

func (w *WorkerPool) Stop() error {
	if atomic.LoadInt32(&w.Status) == STATUS_STOP {
		return ErrPoolIsStoped
	}
	w.StopChan <- struct{}{}
	atomic.StoreInt32(&w.Status, STATUS_STOP)
	return nil
}
