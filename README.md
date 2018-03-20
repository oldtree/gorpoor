# gorpoor goroutine pool

>  一个 goroutine 协程池的实现 ，所有的操作都是预先分配好可以执行的goroutine,避免在连续的实际任务执行中启动新的协程所带来的损耗

>  a goroutine pool implements

>  使用三种方式实现一个可用的goroutine pool：
>  * 一个全局的task队列，然后每一个worker使用注册的回调函数来从全局获取任务并阻塞执行，任务的分发依靠go runtime的调度机制，每一个worker不用维持自己的
>  task队列这样从 主任务 到 worker任务就只有两层，比较简单，worker是不可选的，worker可以执行新的任务时，就必须执行，没有可以做屏蔽的东西（不支持将一个
>  不想去执行的task重新塞入主任务的task队列中）；这样只要投送任务，直至任务队列被填满，但是worker却是有限的
    
    git:(develop*) $ go test -bench .
    goos: darwin
    goarch: amd64
    pkg: github.com/FlyCynomys/tools/gorpoor/v1
    Benchmark_work_10_10-4           1000000              2221 ns/op
    Benchmark_work_10_100-4          2000000              1737 ns/op
    Benchmark_work_100_100-4         2000000               965 ns/op
    Benchmark_work_100_1000-4        2000000               517 ns/op
    PASS
    ok      github.com/FlyCynomys/tools/gorpoor/v1  28.155s


>  * 一个全局的task队列，然后每一个worker在task来到时，被主任务取出并分配任务去执行，执行完结束后使用注册的回调函数来将自己重新入队列，这样worker不断地取出
>  执行的循环，这样只在worker侧保存一个正在执行的task，调度将是齿轮式的，投送任务通过 主任务 来进行,也只是分两层的 主任务 和 worker
>  * 一个全局的task队列,来接受任务，每一个worker也有自己本地的任务队列，每一个worker维持一定数量的goroutine,从worker本地的队列取出任务并执行，并实时更新
>  worker本身的状态，在主任务中可以通过数组来保存worker队列，也可以使用一个buffer channel来保存，通过不断地更新数组状态或者队列的取出放入操作选择合适的
>  worker来派送task去worker本地的队列，当有新任务进来worker的本地队列，或者worker执行完一个task都更新自己的状态，这样就有三层了
>  main-task---->worker-task--->real-goroutine
