package v3

import (
	//"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var result int64
var wg sync.WaitGroup

type MyTask struct {
	Id int64
}

func (m *MyTask) Exec() error {
	atomic.AddInt64(&result, m.Id)
	wg.Done()
	return nil
}

func Test_process(t *testing.T) {
	w := new(WorkerPool)
	w.Init(10, 10)
	go w.Start()
	time.Sleep(time.Second * 2)
	w.Stop()
}

func Test_addtask(t *testing.T) {
	w := new(WorkerPool)
	w.Init(10, 10)
	go w.Start()

	for index := 0; index < 1000; index++ {
		wg.Add(1)
		w.AddTask(&MyTask{Id: int64(index)})
	}
	wg.Wait()
	//log.Println("result : ", result)
	w.Stop()
}

func Benchmark_work_10_10(b *testing.B) {
	w := new(WorkerPool)
	w.Init(10, 10)
	go w.Start()

	for index := 0; index < b.N; index++ {
		wg.Add(1)
		w.AddTask(&MyTask{Id: int64(index)})
	}
	wg.Wait()
	//log.Println("result : ", result)
	w.Stop()
}

func Benchmark_work_10_100(b *testing.B) {
	w := new(WorkerPool)
	w.Init(10, 100)
	go w.Start()
	for index := 0; index < b.N; index++ {
		wg.Add(1)
		w.AddTask(&MyTask{Id: int64(index)})
	}
	wg.Wait()
	//log.Println("result : ", result)
	w.Stop()
}

func Benchmark_work_100_100(b *testing.B) {
	w := new(WorkerPool)
	w.Init(100, 100)
	go w.Start()
	for index := 0; index < b.N; index++ {
		wg.Add(1)
		w.AddTask(&MyTask{Id: int64(index)})
	}
	wg.Wait()
	//log.Println("result : ", result)
	w.Stop()
}

func Benchmark_work_100_1000(b *testing.B) {
	w := new(WorkerPool)
	w.Init(10, 1000)
	go w.Start()

	for index := 0; index < b.N; index++ {
		wg.Add(1)
		w.AddTask(&MyTask{Id: int64(index)})
	}
	wg.Wait()
	//log.Println("result : ", result)
	w.Stop()
}

func Benchmark_work_100_10000(b *testing.B) {
	w := new(WorkerPool)
	w.Init(10, 10000)
	go w.Start()
	for index := 0; index < b.N; index++ {
		wg.Add(1)
		w.AddTask(&MyTask{Id: int64(index)})
	}
	wg.Wait()
	//log.Println("result : ", result)
	w.Stop()
}

func Benchmark_work_1000_10000(b *testing.B) {
	w := new(WorkerPool)
	w.Init(100, 10000)
	go w.Start()

	for index := 0; index < b.N; index++ {
		wg.Add(1)
		w.AddTask(&MyTask{Id: int64(index)})
	}
	wg.Wait()
	//log.Println("result : ", result)
	w.Stop()
}
