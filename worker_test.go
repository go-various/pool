package pool

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var i int32

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
	atomic.AddInt32(&i, 1)
}

func factory() func(i interface{}) (interface{}, error) {
	return func(i interface{}) (interface{}, error) {
		if strings.Contains(i.(string), "subject 5") {
			//time.Sleep(time.Second * 15)
		}
		return i.(string) + " factory executed!", nil
	}
}
func TestNewWorker(t *testing.T) {
	logger := hclog.Default()
	logger.SetLevel(hclog.Debug)
	ctx, _ := context.WithCancel(context.Background())
	worker := NewWorker("worker-1", ctx, factory, logger)

	sub := NewSubject("worker test")
	sub.Observer(NewReader("reader"))

	go worker.Start()
	now := time.Now()
	t.Log(now)
	for i := 0; i < 10000; i++ {
		if err := worker.Input(sub); err != nil {
			t.Fatal(err)
			return
		}
	}
	t.Log(time.Now().Sub(now))
	time.AfterFunc(time.Second*5, func() {
		//worker.Stop()
	})
	t.Log(i)
	for i < 10000 {

	}
	t.Log(i)
}
