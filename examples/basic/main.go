package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Daniel-Sogbey/obs/obs"
)

func main() {
	//root tracker
	ctx := obs.With(context.Background(), "main")
	go worker(ctx, 1)
	go worker(ctx, 2)

	time.Sleep(3 * time.Second)
	printSnapshot()
}

func worker(parent context.Context, id int) {
	ctx := obs.With(parent, fmt.Sprintf("worker %d", id))
	t := obs.FromContext(ctx)
	defer t.Done()

	time.Sleep(time.Duration(id) * time.Second)
	_ = t
}

func printSnapshot() {

	snapshots := obs.Snapshot()

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Id < snapshots[j].Id
	})

	fmt.Println("------- snapshot -------")
	for _, t := range snapshots {
		fmt.Printf("ID: %d Name: %s Parent: %d State: %s Duration: %v\n", t.Id, t.Name, t.ParentId, t.State, t.Duration)
	}
}
