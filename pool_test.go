package pool

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"runtime"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	logger := hclog.New(&hclog.LoggerOptions{Name: "main-pool", Level: hclog.Info})
	ctx, _ := context.WithCancel(context.Background())
	total := 20000000
	pool := NewWorkerPool("main-pool", ctx, logger)
	runtime.GOMAXPROCS(6)
	logger.Info("start")
	//pool.NewWorker(factory)
	//pool.NewWorker(factory)
	//pool.NewWorker(factory)
	//pool.NewWorker(factory)
	pool.AddWorker(new("worker-1", context.Background(), logger))
	pool.AddWorker(new("worker-2", context.Background(), logger))
	pool.AddWorker(new("worker-3", ctx, logger))
	pool.AddWorker(new("worker-4", ctx, logger))
	pool.AddWorker(new("worker-5", ctx, logger))
	pool.AddWorker(new("worker-6", ctx, logger))

	pool.StartWorkers()
	go pool.Start()
	now := time.Now()
	for i := 0; i < total; i++ {
		sub := NewSubject("worker test")
		sub.Observer(NewReader("reader"))
		pool.Input(sub)
	}

	for {
		time.Sleep(time.Millisecond * 1)
		if i >= int32(total) {
			break
		}
	}
	t.Log(time.Now().Sub(now))
	t.Log(i)
	pool.Shutdown()
	<-pool.Running()
	//time.Sleep(time.Second)
}

func new(name string, ctx context.Context, logger hclog.Logger) *Worker {
	worker := NewWorker(name, ctx, factory, logger)
	return worker
}
