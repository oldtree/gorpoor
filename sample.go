package gorpoor

/*
import (
	//"fmt"
	"sync"
)

type TASK int

func GoPool() {
	//starttime := time.Now()
	var taskChan = make(chan TASK, 1000)

	var wg sync.WaitGroup
	for i := 1; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for _ = range taskChan {
				//fmt.Printf("task [%d] info : %d \n", id, value)
			}
		}(i)
	}
	for i := 0; i < 10000000; i++ {
		taskChan <- TASK(i)
	}
	close(taskChan)
	wg.Wait()

	//fmt.Printf("pool exec time : %f \n", time.Now().Sub(starttime).Seconds())
}

func GoNoPool() {
	//starttime := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 10000000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			//fmt.Printf("task [%v] info : %v \n", id, id)
		}(i)
	}

	wg.Wait()

	//fmt.Printf("no pool exec time : %f \n", time.Now().Sub(starttime).Seconds())
}
*/
