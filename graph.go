package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type TaskGraph struct {
	rootId      int64
	readyChanel chan int64
	isRunning   bool
	nodeCount   int64
	doneCount   atomic.Int64

	NodeMap    map[int64]*TaskNode
	NodeError  map[int64]error
	IncreaseId atomic.Int64
	Once       sync.Once
}

func NewTaskGraph() *TaskGraph {
	graph := &TaskGraph{
		nodeCount:   0,
		doneCount:   atomic.Int64{},
		readyChanel: make(chan int64, 1024),
		NodeMap:     make(map[int64]*TaskNode),
		NodeError:   make(map[int64]error),
	}
	graph.doneCount.Store(0)
	graph.IncreaseId.Store(0)

	graph.rootId = graph.AddNode(func(ctx context.Context) error {
		return nil
	}, time.Second)
	graph.readyChanel <- graph.rootId
	return graph
}

func (g *TaskGraph) addInDepth(id int64) {
	g.NodeMap[id].inDepth.Add(1)
}

func (g *TaskGraph) reduceInDepth(id int64) {
	g.NodeMap[id].inDepth.Add(-1)
}

func (g *TaskGraph) AddEdge(parentId, childId int64) {
	parentNode := g.NodeMap[parentId]
	parentNode.Children = append(parentNode.Children, childId)
	g.addInDepth(childId)
}

func (g *TaskGraph) AddNode(node NodeFunc, timeout time.Duration) int64 {
	// IncreaseId cas add
	g.IncreaseId.Add(1)

	taskNode := &TaskNode{
		id:       g.IncreaseId.Load(),
		inDepth:  atomic.Int64{},
		RunFunc:  node,
		Timeout:  timeout,
		Children: make([]int64, 0),
	}
	taskNode.inDepth.Store(0)
	g.NodeMap[taskNode.id] = taskNode
	g.NodeError[taskNode.id] = nil
	g.nodeCount++
	if g.nodeCount != 1 {
		g.AddEdge(g.rootId, taskNode.id)
	}
	return taskNode.id
}

func (g *TaskGraph) stop() {
	g.Once.Do(func() {
		g.isRunning = false
		close(g.readyChanel)
		println("done graph")
	})
}

func (g *TaskGraph) doneNode() {
	g.doneCount.Add(1)
	if g.doneCount.Load() == g.nodeCount {
		g.stop()
	}
}

// 添加新的方法用于检测环
func (g *TaskGraph) detectCycle() error {
	visited := make(map[int64]bool)
	recursionStack := make(map[int64]bool)

	var dfs func(nodeId int64) bool
	dfs = func(nodeId int64) bool {
		visited[nodeId] = true
		recursionStack[nodeId] = true

		for _, childId := range g.NodeMap[nodeId].Children {
			if !visited[childId] {
				if dfs(childId) {
					return true
				}
			} else if recursionStack[childId] {
				return true
			}
		}

		recursionStack[nodeId] = false
		return false
	}

	for nodeId := range g.NodeMap {
		if !visited[nodeId] {
			if dfs(nodeId) {
				return errors.New("ring detected, dag invalid")
			}
		}
	}

	return nil
}

func (g *TaskGraph) Run(ctx context.Context) error {
	// 在运行前检测环
	if err := g.detectCycle(); err != nil {
		logrus.Errorln("detectCycle error:", err)

		return err
	}

	g.isRunning = true
	// first check haven't cycle
	for g.isRunning {
		nodeId, ok := <-g.readyChanel
		if !ok {
			return nil
		}
		go func() {
			node := g.NodeMap[nodeId]
			node.Once.Do(func() {
				g.NodeError[nodeId] = node.RunFunc(ctx)
				g.doneNode()
			})
			for _, childId := range node.Children {
				g.reduceInDepth(childId)
				if g.NodeMap[childId].inDepth.Load() == 0 {
					g.readyChanel <- childId
				}
			}
		}()
	}
	return nil
}
