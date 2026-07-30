# Scenario

**Feature**: wrk --bring works when cwd is inside a sub-module directory (not the repo root)

```
# consumer cwd = consumer/go-pkgs/ (the dir with go.mod) -> wrk --bring -> success
consumer (cwd = go-pkgs/) + dep -> wrk --bring -> external worktree + replace
```

## Preconditions

- Git and Go must be available.
- Consumer repo root has NO go.mod; `go-pkgs/go.mod` requires dep.
- Consumer cwd is `consumer/go-pkgs/` (NOT the repo root).

## Steps

1. Create consumer git repo with `go-pkgs/go.mod` requiring dep.
2. Create dep git repo with root `go.mod`.
3. Run `wrk --bring <dep>` with cwd = `consumer/go-pkgs/`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

import "path/filepath"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	modDir := filepath.Join(consumer, "go-pkgs")
	mkdirAll(t, modDir)
	writeFile(t, filepath.Join(modDir, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire "+bringDepModulePath+" v0.0.0\n")
	writeFile(t, filepath.Join(modDir, "main.go"), "package main\n")
	runGitIsolated(t, consumer, "add", ".")
	runGitIsolated(t, consumer, "commit", "-m", "add sub-module")

	depPath := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	// cwd is inside go-pkgs/, not the repo root
	req.RepoDir = filepath.Join(consumer, "go-pkgs")
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.ConsumerModDir = modDir
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", depPath}
	return nil
}
```