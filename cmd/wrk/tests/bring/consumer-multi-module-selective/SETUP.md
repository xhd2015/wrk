# Scenario

**Feature**: wrk --bring replaces only in consumer modules that actually depend on the dep

```
# consumer has go-pkgs/go.mod (requires dep) and tools/go.mod (does NOT require dep) -> wrk --bring -> replace only in go-pkgs/
consumer (go-pkgs/ requires dep, tools/ does not) + dep -> wrk --bring -> replace in go-pkgs/go.mod only
```

## Preconditions

- Git and Go must be available.
- Consumer repo root has NO go.mod; `go-pkgs/go.mod` requires dep, `tools/go.mod` does not.
- Consumer cwd is the repo root.

## Steps

1. Create consumer git repo with `go-pkgs/go.mod` (requires dep) and `tools/go.mod` (no dep require).
2. Create dep git repo with root `go.mod` (`module example.com/dep`).
3. Run `wrk --bring <dep>` from the consumer repo root.

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
	modDir1 := filepath.Join(consumer, "go-pkgs")
	modDir2 := filepath.Join(consumer, "tools")
	mkdirAll(t, modDir1)
	mkdirAll(t, modDir2)
	writeFile(t, filepath.Join(modDir1, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire "+bringDepModulePath+" v0.0.0\n")
	writeFile(t, filepath.Join(modDir1, "main.go"), "package main\n")
	// tools has NO dep require.
	writeFile(t, filepath.Join(modDir2, "go.mod"), "module example.com/consumer-tools\n\ngo 1.22\n")
	writeFile(t, filepath.Join(modDir2, "tool.go"), "package tools\n")
	runGitIsolated(t, consumer, "add", ".")
	runGitIsolated(t, consumer, "commit", "-m", "add sub-modules")

	depPath := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.ConsumerModDir = modDir1
	req.Args = []string{"--bring", depPath}
	return nil
}
```