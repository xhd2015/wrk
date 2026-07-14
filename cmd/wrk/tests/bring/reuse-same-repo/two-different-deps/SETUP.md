# Scenario

**Feature**: two different dep mains never false-reuse each other's external worktrees

```
# bring mydep1 then bring mydep2 (distinct main repos / module paths)
# each gets its own external/{basename}-main-{date}; no reuse warning for dep2
consumer (require dep1+dep2)
  -> wrk --bring mydep1 -> external/mydep1-…
  -> wrk --bring mydep2 -> external/mydep2-… (new; not mydep1 path)
```

## Steps

1. Create two dep repos (`mydep1`/`example.com/dep1`, `mydep2`/`example.com/dep2`) and a consumer requiring both.
2. `--bring mydep1` once (precondition).
3. Run `wrk --bring mydep2` via `Run`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureBringReuseHelpersUsed()

	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runBringGo(t, consumer, "mod", "edit", "-require=example.com/dep1@v0.0.0")
	runBringGo(t, consumer, "mod", "edit", "-require=example.com/dep2@v0.0.0")
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks consumer: %v", err)
	}

	dep1 := filepath.Join(req.WorkRoot, "mydep1")
	initGitRepoOnMain(t, dep1)
	writeFile(t, filepath.Join(dep1, "go.mod"), "module example.com/dep1\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dep1, "dep.go"), "package dep1\n")
	runGitIsolated(t, dep1, "add", "go.mod", "dep.go")
	runGitIsolated(t, dep1, "commit", "-m", "add dep1")
	dep1, err = filepath.EvalSymlinks(dep1)
	if err != nil {
		t.Fatalf("eval symlinks dep1: %v", err)
	}

	dep2 := filepath.Join(req.WorkRoot, "mydep2")
	initGitRepoOnMain(t, dep2)
	writeFile(t, filepath.Join(dep2, "go.mod"), "module example.com/dep2\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dep2, "dep.go"), "package dep2\n")
	runGitIsolated(t, dep2, "add", "go.mod", "dep.go")
	runGitIsolated(t, dep2, "commit", "-m", "add dep2")
	dep2, err = filepath.EvalSymlinks(dep2)
	if err != nil {
		t.Fatalf("eval symlinks dep2: %v", err)
	}

	first := runWrkWithArgs(t, req, consumer, "--bring", dep1)
	want1 := bringExternalWorktreePath(consumer, "mydep1", "main", 0)
	if first != want1 {
		t.Fatalf("first --bring mydep1: expected %q, got %q", want1, first)
	}
	// First --bring runs tidy and may drop unused requires; re-pin dep2 for the second bring.
	runBringGo(t, consumer, "mod", "edit", "-require=example.com/dep2@v0.0.0")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep2
	req.ExternalWtDir = want1 // prior other-dep path (must not be reused)
	req.Args = []string{"--bring", dep2}
	return nil
}
```
