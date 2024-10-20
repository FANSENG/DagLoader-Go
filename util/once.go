package util

import (
	"sync"
	"sync/atomic"
	"time"
)

type Once struct {
	done uint32
	m    sync.Mutex
}

func (o *Once) Do(f func()) {
	if atomic.LoadUint32(&o.done) == 0 {
		o.doSlow(f)
	}
}

func (o *Once) DoWithTimeout(f func(), timeout time.Duration) bool {
	if atomic.LoadUint32(&o.done) == 0 {
		return o.doSlowWithTimeout(f, timeout)
	}
	return true
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}

func (o *Once) doSlowWithTimeout(f func(), timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		o.m.Lock()
		defer o.m.Unlock()
		defer close(done)
		if o.done == 0 {
			defer atomic.StoreUint32(&o.done, 1)
			f()
		}
	}()

	select {
	case <-done: // 完成任务
		return true
	case <-time.After(timeout): // 超时
		return false
	}
}
