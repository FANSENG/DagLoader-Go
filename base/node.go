package base

import (
	"context"
	"sync/atomic"
	"time"

	"fs1n/dag/util"
)

type NodeFunc func(context.Context) error

type TaskNode struct {
	id      int64
	inDepth atomic.Int64

	RunFunc  NodeFunc
	Once     util.Once
	Timeout  time.Duration
	Children []int64
}
