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
	ErrPoolIsRunning   = errors.New("worker pool is running")
	ErrPoolInitFailed  = errors.New("worker pool init failed")
	ErrPoolAlreadyStop = errors.New("worker pool is already stop")
)

type TASK func()

type Worker interface {
	Init(chan struct{}) error
	Start() error
	Stop() error
	Rest() error
	GetBusy(t TASK) error
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

func (w *WorkerPool) Init() error {
	atomic.StoreInt32(&w.Status, STATUS_START)

	return ErrPoolInitFailed
}

func (w *WorkerPool) Start() error {
	if atomic.LoadInt32(&w.Status) == STATUS_RUNNING {
		return ErrPoolIsRunning
	}
	defer func() {
		atomic.StoreInt32(&w.Status, STATUS_STOP)
	}()
WORKLOOP:
	for {
		select {
		case newT := <-w.Tasklist:
			w := <-w.WorkerList
			w.GetBusy(newT)
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
		return ErrPoolAlreadyStop
	}
	w.StopChan <- struct{}{}
	return nil
}
