package main

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type NodeFunc func(context.Context) error

type TaskNode struct {
	id      int64
	inDepth atomic.Int64

	RunFunc  NodeFunc
	Once     sync.Once
	Timeout  time.Duration
	Children []int64
}
