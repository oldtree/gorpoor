package v1

import (
	"fmt"
	"sync/atomic"
	"testing"
)

var counter int64 = 0

var f = func() interface{} {
	atomic.AddInt64(&counter, 1)
	return nil
}

func TestGorPoor(t *testing.T) {
	var gp = NewWorkerPool(4, 8, BoxGenericWorkor)
	var err error
	for index := 0; index < 16; index++ {
		err = gp.Accept(Task(f))
		if err != nil {
			t.Errorf("exec func [%p] failed ", f)
		}
	}
	/*err = gp.Stop()
	if err != nil {
		t.Errorf("stop gorpoor [%v] failed ", err)
	}*/
}

func BenchmarkGorPoorTask(b *testing.B) {
	var gp = NewWorkerPool(8, 100, BoxGenericWorkor)
	var err error
	for i := 0; i < b.N; i++ {
		err = gp.Accept(Task(f))
		if err != nil {
			b.Errorf("exec func [%p] failed ", f)
		}
	}
	/*err = gp.Stop()
	if err != nil {
		b.Errorf("stop gorpoor [%v] failed ", err)
	}
	fmt.Printf("total task exec : %d \n", counter)
	*/
	fmt.Printf("total task exec : %d \n", counter)
}

func OneSample() {
	var gp = NewWorkerPool(8, 100, BoxGenericWorkor)
	for index := 0; index < 1000; index++ {
		gp.Accept(Task(f))
	}
	fmt.Printf("total task exec : %d \n", counter)
}

func BenchmarkGorPoorWorker(b *testing.B) {
	for i := 0; i < b.N; i++ {
		OneSample()
	}
}
