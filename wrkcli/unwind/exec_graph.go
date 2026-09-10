package unwind

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
)

// graphExecOpts configures concurrent DAG apply (go build Builder.Do style).
type graphExecOpts struct {
	DryRun  bool
	OnStart func(a *Action)
	OnDone  func(a *Action, err error)
}

type execNode struct {
	action   *Action
	priority int
	pending  int
	triggers []*execNode
	failed   *execNode
	isRoot   bool
}

type readyHeap []*execNode

func (h readyHeap) Len() int { return len(h) }
func (h readyHeap) Less(i, j int) bool {
	if h[i].priority != h[j].priority {
		return h[i].priority > h[j].priority
	}
	return h[i].action.ID < h[j].action.ID
}
func (h readyHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *readyHeap) Push(x interface{}) { *h = append(*h, x.(*execNode)) }
func (h *readyHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// runActionGraph schedules actions by Deps with up to jobs workers.
// Fail-fast: first error cancels the context; dependents are marked skipped.
func runActionGraph(ctx context.Context, actions []*Action, jobs int, opts graphExecOpts, act func(context.Context, *Action) error) error {
	if len(actions) == 0 {
		return nil
	}
	if jobs < 1 {
		jobs = 1
	}
	if opts.DryRun {
		jobs = 1
	}

	byID := make(map[string]*execNode, len(actions)+1)
	nodes := make([]*execNode, 0, len(actions))
	for _, a := range actions {
		if a == nil || a.ID == "" {
			continue
		}
		n := &execNode{action: a}
		byID[a.ID] = n
		nodes = append(nodes, n)
	}
	if len(nodes) == 0 {
		return nil
	}

	rootAct := &Action{ID: "exec:root", Mode: "exec-root"}
	root := &execNode{action: rootAct, isRoot: true}
	byID[rootAct.ID] = root

	post := topoPostOrder(nodes, byID)
	for i, n := range post {
		n.priority = i
	}
	root.priority = len(post) + 1

	for _, n := range nodes {
		seen := map[string]bool{}
		for _, d := range n.action.Deps {
			dep := byID[d]
			if dep == nil || seen[d] {
				continue
			}
			seen[d] = true
			dep.triggers = append(dep.triggers, n)
			n.pending++
		}
		// Every real node triggers the synthetic root.
		n.triggers = append(n.triggers, root)
		root.pending++
	}

	var (
		execMu   sync.Mutex
		readyMu  sync.Mutex
		ready    readyHeap
		readySem = make(chan bool, len(nodes)+1)
		wg       sync.WaitGroup
		firstErr error
	)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	pushReady := func(n *execNode) {
		readyMu.Lock()
		heap.Push(&ready, n)
		readyMu.Unlock()
		readySem <- true
	}

	for _, n := range nodes {
		if n.pending == 0 {
			pushReady(n)
		}
	}

	handle := func(n *execNode) {
		if n.isRoot {
			close(readySem)
			return
		}
		if opts.OnStart != nil {
			opts.OnStart(n.action)
		}
		var err error
		if n.failed != nil {
			err = fmt.Errorf("skipped: dependency %s failed", n.failed.action.ID)
		} else {
			select {
			case <-ctx.Done():
				err = ctx.Err()
			default:
				err = act(ctx, n.action)
			}
		}
		if opts.OnDone != nil {
			opts.OnDone(n.action, err)
		}

		execMu.Lock()
		defer execMu.Unlock()
		if err != nil && firstErr == nil && n.failed == nil {
			firstErr = err
			n.failed = n
			cancel()
		}
		fail := n.failed
		for _, t := range n.triggers {
			if fail != nil && t.failed == nil {
				t.failed = fail
			}
			t.pending--
			if t.pending == 0 {
				pushReady(t)
			}
		}
	}

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				_, ok := <-readySem
				if !ok {
					return
				}
				readyMu.Lock()
				if ready.Len() == 0 {
					readyMu.Unlock()
					continue
				}
				n := heap.Pop(&ready).(*execNode)
				readyMu.Unlock()
				handle(n)
			}
		}()
	}
	wg.Wait()
	return firstErr
}

func topoPostOrder(nodes []*execNode, byID map[string]*execNode) []*execNode {
	seen := map[string]bool{}
	var out []*execNode
	var walk func(*execNode)
	walk = func(n *execNode) {
		if n == nil || n.isRoot || seen[n.action.ID] {
			return
		}
		seen[n.action.ID] = true
		for _, d := range n.action.Deps {
			walk(byID[d])
		}
		out = append(out, n)
	}
	for _, n := range nodes {
		walk(n)
	}
	return out
}

// resolveJobs maps UnwindFlags.Jobs: 0 → gomaxprocs, negative → 1.
func resolveJobs(jobs, gomaxprocs int) int {
	if jobs < 0 {
		return 1
	}
	if jobs == 0 {
		if gomaxprocs < 1 {
			return 1
		}
		return gomaxprocs
	}
	return jobs
}
