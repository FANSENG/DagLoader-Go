package main

import (
	"context"
	"time"
)

func main() {
	graph := NewTaskGraph()
	nodeId2 := graph.AddNode(func(ctx context.Context) error {
		println("2")
		return nil
	}, time.Second)
	nodeId3 := graph.AddNode(func(ctx context.Context) error {
		println("3")
		return nil
	}, time.Second)
	nodeId4 := graph.AddNode(func(ctx context.Context) error {
		println("4")
		return nil
	}, time.Second)
	nodeId5 := graph.AddNode(func(ctx context.Context) error {
		println("5")
		return nil
	}, time.Second)

	graph.AddEdge(nodeId5, nodeId2)
	graph.AddEdge(nodeId2, nodeId3)
	graph.AddEdge(nodeId3, nodeId4)
	graph.AddEdge(nodeId4, nodeId5)
	_ = graph.Run(context.Background())
}

// go run main.go graph.go node.go
