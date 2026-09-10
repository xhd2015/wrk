package unwind

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunActionGraphDiamondConcurrency(t *testing.T) {
	//   a
	//  / \
	// b   c
	//  \ /
	//   d
	actions := []*Action{
		{ID: "a", Mode: "nop", Lane: "x"},
		{ID: "b", Mode: "nop", Lane: "x", Deps: []string{"a"}},
		{ID: "c", Mode: "nop", Lane: "y", Deps: []string{"a"}},
		{ID: "d", Mode: "nop", Lane: "x", Deps: []string{"b", "c"}},
	}
	var (
		mu      sync.Mutex
		started = map[string]time.Time{}
		inBC    atomic.Int32
		sawBoth = make(chan struct{})
		once    sync.Once
		release = make(chan struct{})
	)

	errCh := make(chan error, 1)
	go func() {
		errCh <- runActionGraph(context.Background(), actions, 4, graphExecOpts{}, func(ctx context.Context, a *Action) error {
			mu.Lock()
			started[a.ID] = time.Now()
			mu.Unlock()
			switch a.ID {
			case "b", "c":
				if inBC.Add(1) == 2 {
					once.Do(func() { close(sawBoth) })
				}
				<-release
			}
			return nil
		})
	}()
	select {
	case <-sawBoth:
		close(release)
	case <-time.After(3 * time.Second):
		close(release)
		t.Fatal("timeout waiting for b and c to overlap")
	}
	err := <-errCh
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if started["d"].IsZero() || started["b"].IsZero() || started["c"].IsZero() {
		t.Fatalf("missing starts: %v", started)
	}
}

func TestRunActionGraphFailFast(t *testing.T) {
	actions := []*Action{
		{ID: "a", Mode: "nop"},
		{ID: "b", Mode: "nop", Deps: []string{"a"}},
		{ID: "c", Mode: "nop", Deps: []string{"a"}},
	}
	var ranB, ranC atomic.Int32
	err := runActionGraph(context.Background(), actions, 2, graphExecOpts{}, func(ctx context.Context, a *Action) error {
		if a.ID == "a" {
			return fmt.Errorf("boom")
		}
		if a.ID == "b" {
			ranB.Add(1)
		}
		if a.ID == "c" {
			ranC.Add(1)
		}
		return nil
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("want boom, got %v", err)
	}
	if ranB.Load()+ranC.Load() != 0 {
		t.Fatalf("dependents should not run after fail-fast, b=%d c=%d", ranB.Load(), ranC.Load())
	}
}

func TestRunActionGraphSerialJobsOne(t *testing.T) {
	actions := []*Action{
		{ID: "a", Mode: "nop"},
		{ID: "b", Mode: "nop"},
	}
	var cur atomic.Int32
	err := runActionGraph(context.Background(), actions, 1, graphExecOpts{}, func(ctx context.Context, a *Action) error {
		if cur.Add(1) != 1 {
			return fmt.Errorf("overlap with jobs=1")
		}
		time.Sleep(20 * time.Millisecond)
		cur.Add(-1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
