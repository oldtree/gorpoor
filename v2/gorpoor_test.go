package v2

import (
	"sync"
	"testing"
	"time"
)

var resultmap = new(sync.Map)

type MyTask struct {
	Id int
}

func (m *MyTask) Exec() error {
	resultmap.Store(m.Id, true)
	return nil
}

func Test_process(t *testing.T) {
	w := new(WorkerPool)
	w.Init(10, 10)
	go w.Start()
	time.Sleep(time.Second * 1)
	w.Stop()
}

func Test_addtask(t *testing.T) {
	w := new(WorkerPool)
	w.Init(10, 10)
	go w.Start()

	for index := 0; index < 1000; index++ {
		w.AddTask(&MyTask{Id: index})
	}
	itf := func(key interface{}, value interface{}) bool {
		return true
	}
	resultmap.Range(itf)
	time.Sleep(time.Second * 2)
	w.Stop()
}

func Benchmark_work_10_10(b *testing.B) {
	w := new(WorkerPool)
	w.Init(10, 10)
	go w.Start()
	defer w.Stop()
	for index := 0; index < b.N; index++ {
		w.AddTask(&MyTask{Id: index})
	}
	itf := func(key interface{}, value interface{}) bool {
		return true
	}
	resultmap.Range(itf)
}

func Benchmark_work_10_100(b *testing.B) {
	w := new(WorkerPool)
	w.Init(10, 100)
	go w.Start()
	defer w.Stop()
	for index := 0; index < b.N; index++ {
		w.AddTask(&MyTask{Id: index})
	}
	itf := func(key interface{}, value interface{}) bool {
		return true
	}
	resultmap.Range(itf)
}

func Benchmark_work_100_100(b *testing.B) {
	w := new(WorkerPool)
	w.Init(100, 100)
	go w.Start()
	defer w.Stop()
	for index := 0; index < b.N; index++ {
		w.AddTask(&MyTask{Id: index})
	}
	itf := func(key interface{}, value interface{}) bool {
		return true
	}
	resultmap.Range(itf)
}

func Benchmark_work_100_1000(b *testing.B) {
	w := new(WorkerPool)
	w.Init(10, 1000)
	go w.Start()
	defer w.Stop()
	for index := 0; index < b.N; index++ {
		w.AddTask(&MyTask{Id: index})
	}
	itf := func(key interface{}, value interface{}) bool {
		return true
	}
	resultmap.Range(itf)
}
