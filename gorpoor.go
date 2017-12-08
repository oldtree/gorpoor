package gorpoor

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	STATUS_START = iota
	STATUS_RUNNING
	STATUS_STOP
	STATUS_WAITING
	STATUS_PAUSE
	STATUS_RESTART
)

var (
	ErrWorkPoolIsRuning = errors.New("work pool is running")
)

type Task func() interface{}
type DeliverTask func() chan Task
type RestWorker func(Worker)
type GenericWorkor func(stopChan chan struct{}) Worker

func BoxGenericWorkor(stopChan chan struct{}) Worker {
	return &BoxWorker{
		StopChan: stopChan,
		QuitW:    make(chan struct{}, 1),
		Status:   STATUS_START,
		WorkerID: time.Now().UnixNano(),
	}
}

type Worker interface {
	Init(DeliverTask, RestWorker) Worker
	Start(DeliverTask, RestWorker) Worker
	Stop() Worker
}

type BoxWorker struct {
	WorkerID int64
	StopChan chan struct{} // broadcast chan ,anyone recive must close this chanel

	QuitW  chan struct{} // quit chan
	Status int32
}

func (s *BoxWorker) Init(getBusyLiving DeliverTask, getBusydie RestWorker) Worker {
	go s.Start(getBusyLiving, getBusydie)
	return s
}

func (s *BoxWorker) Start(busy DeliverTask, rest RestWorker) Worker {
	atomic.StoreInt32(&s.Status, STATUS_START)
WORKPOOL:
	for {
		select {
		case letbusy := <-busy():
			atomic.StoreInt32(&s.Status, STATUS_RUNNING)
			letbusy()
			//rest(s)
			atomic.StoreInt32(&s.Status, STATUS_WAITING)
		case _, ok := <-s.StopChan:
			if !ok {
				fmt.Println("work loop is stop")
			} else {
				close(s.StopChan)
			}
			break WORKPOOL
		case <-s.QuitW:
			fmt.Println("worker [%d] is quit from pool", s.WorkerID)
			close(s.QuitW)
			break WORKPOOL
		}
	}
	return nil
}

func (s *BoxWorker) Stop() Worker {
	if atomic.LoadInt32(&s.Status) == STATUS_STOP {
		return nil
	}
	s.QuitW <- struct{}{}
	atomic.StoreInt32(&s.Status, STATUS_STOP)
	return s
}

type WorkGoroutinePool struct {
	WaitGroup sync.WaitGroup

	TaskG    chan Task
	WorkerG  chan Worker
	StopChan chan struct{}

	IsRunning bool
}

func NewWorkerPool(WorkorNumber int, TaskQueue int, gw GenericWorkor) *WorkGoroutinePool {
	wp := &WorkGoroutinePool{}
	wp.init(WorkorNumber, TaskQueue, gw)
	wp.StopChan = make(chan struct{}, 1)
	go wp.Start()
	return wp
}

func (wp *WorkGoroutinePool) init(WorkorNumber int, TaskQueue int, gw GenericWorkor) {
	wp.WorkerG = make(chan Worker, WorkorNumber)
	wp.TaskG = make(chan Task, TaskQueue)
	for index := 0; index < WorkorNumber; index++ {
		work := gw(wp.StopChan)
		work.Init(wp.makeWorkerBusy, wp.makeWorkerRest)
		wp.WorkerG <- work
	}
}

func (wp *WorkGoroutinePool) makeWorkerBusy() chan Task {
	return wp.TaskG
}

func (wp *WorkGoroutinePool) makeWorkerRest(worker Worker) {
	wp.WorkerG <- worker
}

func (wp *WorkGoroutinePool) Start() error {
	defer func() {
		wp.IsRunning = false
		fmt.Println("workor loop is exit")
	}()
	if wp.IsRunning == true {
		return ErrWorkPoolIsRuning
	}
	wp.IsRunning = true
	println("pool start")
WORKPOOL:
	for {
		select {
		case _, ok := <-wp.StopChan:
			if !ok {
				fmt.Println("work loop is stop")
			} else {
				close(wp.StopChan)
			}
			break WORKPOOL
		}
	}
	return nil
}

func (wp *WorkGoroutinePool) Stop() error {
	if wp.IsRunning {
		wp.StopChan <- struct{}{}
	} else {
		fmt.Println("work pool is stop")
	}
	return nil
}

func (wp *WorkGoroutinePool) Accept(t Task) error {
	//if wp.IsRunning {
	println("start send task")
	wp.TaskG <- t
	println("task send ok")
	//}
	return nil
}

/*
//cause the go-1.9.2 goreturns cannot fmt this file ,than disable this below
type GPool = *WorkGoroutinePool

func NewGPool() GPool {
	g:= GPool(new(WorkGoroutinePool))
	return g
}

func (g GPool) AddTask(t Task) {
	g.Accept(t)
}

func (g GPool) StopTask() {
	g.Stop()
}
*/
