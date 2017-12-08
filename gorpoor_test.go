package gorpoor

import (
	"testing"
	"time"
)

var f = func() interface{} {
	println(time.Now().String())
	return nil
}

func TestGorPoor(t *testing.T) {
	var gp = NewWorkerPool(4, 8, BoxGenericWorkor)
	var err error
	println("into loop")
	for index := 0; index < 16; index++ {
		err = gp.Accept(Task(f))
		if err != nil {
			t.Errorf("exec func [%p] failed ", f)
		}
	}
	time.Sleep(time.Second * 4)
	err = gp.Stop()
	if err != nil {
		t.Errorf("stop gorpoor [%v] failed ", err)
	}
}

func BenchmarkGorPoor(b *testing.B) {
	var gp = NewWorkerPool(20, 100000, BoxGenericWorkor)
	var err error
	for i := 0; i < b.N; i++ {
		err = gp.Accept(Task(f))
		if err != nil {
			b.Errorf("exec func [%p] failed ", f)
		}
	}
	time.Sleep(time.Minute * 4)
	err = gp.Stop()
	if err != nil {
		b.Errorf("stop gorpoor [%v] failed ", err)
	}
}
