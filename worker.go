package pool

import (
	"context"
	"errors"
	"github.com/hashicorp/go-hclog"
	"runtime"
	"sync/atomic"
)

var ErrWorkFull = errors.New("worker chan is full")
var ErrWorkStopped = errors.New("worker stopped")

//任务结构体
type Worker struct {
	factory   Factory //任务执行方法
	name      string
	running   chan bool
	stopChan  chan bool     //用于pool自身的shutdown方法关闭
	inputChan chan *subject //数据输入chan
	logger    hclog.Logger
	stopped   bool
	deals     uint64
}

func NewWorker(name string, ctx context.Context, factory Factory, logger hclog.Logger) *Worker {
	return &Worker{
		name:      name,
		factory:   factory,
		stopChan:  make(chan bool, 1),
		logger:    logger.Named(name),
		inputChan: make(chan *subject, 16384),
		running:   make(chan bool, 1),
	}
}

func (m *Worker) Running() <-chan bool {
	return m.running
}

func (m *Worker) ChanSize() int {
	return len(m.inputChan)
}

func (m *Worker) Input(s *subject) error {
	if m.stopped {
		return ErrWorkStopped
	}

	if len(m.inputChan) == cap(m.inputChan) {
		return ErrWorkFull
	}
	m.inputChan <- s
	return nil
}
func (m *Worker) Stop() {
	if m.logger.IsTrace() {
		m.logger.Trace("stopping...")
	}
	m.stopChan <- true
}

func (m *Worker) Start() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer func() {
		if m.logger.IsTrace() {
			m.logger.Trace("exited")
		}
		m.logger.Info("worker deals", "total", m.deals)
	}()
	isRunning := true
	for isRunning || m.ChanSize() > 0 {
		select {
		case s := <-m.inputChan:
			if m.logger.IsTrace() {
				m.logger.Trace("started")
			}
			atomic.AddUint64(&m.deals, 1)
			result, err := m.factory()(s.data)
			s.updateContext(result, err)

			if m.logger.IsTrace() {
				m.logger.Trace("finished")
			}
		case <-m.stopChan:
			isRunning = false
			m.stopped = true
		}
	}
	m.cleanup()

}

func (m *Worker) cleanup() {
	if m.logger.IsTrace() {
		m.logger.Trace("cleanup...")
	}

	close(m.stopChan)
	close(m.inputChan)
	close(m.running)
}
