package v1

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	STATUS_START = iota
	STATUS_RUNNING
	STATUS_STOP
	STATUS_WAITING
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
	StopChan chan struct{} // broadcast chan ,anyone receive must close this chanel
	QuitW    chan struct{} // quit chan
	Status   int32
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
				log.Println("worker is stop")
			} else {
				close(s.StopChan)
				log.Println("worker is stop")
			}
			atomic.StoreInt32(&s.Status, STATUS_STOP)
			break WORKPOOL
		case <-s.QuitW:
			log.Printf("worker [%d] is quit from pool \n", s.WorkerID)
			close(s.QuitW)
			atomic.StoreInt32(&s.Status, STATUS_STOP)
			break WORKPOOL
		}
	}
	return nil
}

func (s *BoxWorker) Stop() Worker {
	if atomic.LoadInt32(&s.Status) == STATUS_STOP {
		return s
	}
	s.QuitW <- struct{}{}
	atomic.StoreInt32(&s.Status, STATUS_STOP)
	return s
}

type WorkerPool struct {
	WaitGroup sync.WaitGroup

	TaskG    chan Task
	WorkerG  chan Worker
	StopChan chan struct{}

	IsRunning bool
}

func NewWorkerPool(WorkorNumber int, TaskQueue int, gw GenericWorkor) *WorkerPool {
	wp := &WorkerPool{}
	wp.Init(WorkorNumber, TaskQueue, gw)
	wp.StopChan = make(chan struct{}, 1)
	go wp.Start()
	return wp
}

func (wp *WorkerPool) Init(WorkorNumber int, TaskQueue int, gw GenericWorkor) {
	if gw == nil {
		gw = BoxGenericWorkor
	}
	wp.WorkerG = make(chan Worker, WorkorNumber)
	wp.TaskG = make(chan Task, TaskQueue)
	for index := 0; index < WorkorNumber; index++ {
		work := gw(wp.StopChan)
		work.Init(wp.makeWorkerBusy, wp.makeWorkerRest)
		wp.WorkerG <- work
	}
	//log.Printf("total [%d] worker is start \n", WorkorNumber)
}

func (wp *WorkerPool) makeWorkerBusy() chan Task {
	return wp.TaskG
}

func (wp *WorkerPool) makeWorkerRest(worker Worker) {
	wp.WorkerG <- worker
}

func (wp *WorkerPool) Start() error {
	defer func() {
		wp.IsRunning = false
	}()
	if wp.IsRunning == true {
		return ErrWorkPoolIsRuning
	}
	wp.IsRunning = true
WORKPOOL:
	for {
		select {
		case _, ok := <-wp.StopChan:
			if !ok {
				log.Println("work pool is stop")
			} else {
				close(wp.StopChan)
			}
			break WORKPOOL
		}
	}
	return nil
}

func (wp *WorkerPool) Stop() error {
	if wp.IsRunning {
		wp.StopChan <- struct{}{}
	} else {
		log.Println("work pool is stop")
	}
	return nil
}

func (wp *WorkerPool) Accept(t Task) error {
	if wp.IsRunning {
		wp.TaskG <- t
		return nil
	}
	return nil
}
