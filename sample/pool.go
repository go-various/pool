package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-various/pool"
	"github.com/hashicorp/go-hclog"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync/atomic"
	"time"
)

var i uint32

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:8000", nil)
	}()
	cpu, err := os.Create("cpu.profile")
	if err != nil {
		log.Fatal(err)
		return
	}
	buf := bufio.NewWriter(cpu)

	mem, _ := os.Create("mem.profile")

	membuf := bufio.NewWriter(mem)
	//if err := pprof.StartCPUProfile(buf); err != nil {
	//	log.Fatal(err)
	//	return
	//}
	//
	//defer func() {
	//	pprof.StopCPUProfile()
	//}()
	logger := hclog.New(&hclog.LoggerOptions{Name: "main-workerPool", Level: hclog.Info})
	ctx, _ := context.WithCancel(context.Background())
	total := 10000000
	workerPool := pool.NewWorkerPool("main-workerPool", ctx, logger)
	runtime.GOMAXPROCS(6)
	logger.Info("start")
	//workerPool.NewWorker(factory)
	//workerPool.NewWorker(factory)
	//workerPool.NewWorker(factory)
	//workerPool.NewWorker(factory)
	workerPool.AddWorker(new("worker-1", context.Background(), logger))
	workerPool.AddWorker(new("worker-2", context.Background(), logger))
	workerPool.AddWorker(new("worker-3", ctx, logger))
	workerPool.AddWorker(new("worker-4", ctx, logger))
	workerPool.AddWorker(new("worker-5", ctx, logger))
	workerPool.AddWorker(new("worker-6", ctx, logger))

	workerPool.StartWorkers()
	go workerPool.Start()
	now := time.Now()
	for i := 0; i < total; i++ {
		sub := pool.NewSubject(fmt.Sprintf("worker test %d", i))
		sub.Observer(NewReader("reader"))
		workerPool.Input(sub)
		time.Sleep(time.Microsecond * 5)
	}
	//
	//for {
	//	time.Sleep(time.Millisecond * 1)
	//	if i >= uint32(total) {
	//		break
	//	}
	//}
	atomic.LoadUint32(&i)
	logger.Info("finished", "since", time.Now().Sub(now))
	workerPool.Shutdown()
	<-workerPool.Running()
	runtime.GC() // GC，获取最新的数据信息

	if err := pprof.WriteHeapProfile(membuf); err != nil {
		log.Fatal(err)
		return
	}
	membuf.Flush()
	buf.Flush()
	time.Sleep(time.Second * 900)
}

type Reader struct {
	name string
}

func NewReader(name string) *Reader {
	return &Reader{
		name: name,
	}
}

func (r *Reader) Update(result interface{}, err error) {
	//fmt.Println("observer", r.name, result, err)
	atomic.AddUint32(&i, 1)
}
func new(name string, ctx context.Context, logger hclog.Logger) *pool.Worker {
	worker := pool.NewWorker(name, ctx, factory, logger)
	return worker
}
func factory(i interface{}) (interface{}, error) {
	if strings.Contains(i.(string), " 5") {
	}
	return i.(string) + " factory executed!", nil
}
