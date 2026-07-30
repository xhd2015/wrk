# Scenario

**Feature**: wrk --bring replaces in all consumer modules that depend on the dep

```
# consumer has go-pkgs/go.mod and tools/go.mod, both requiring dep -> wrk --bring -> replace in both
consumer (go-pkgs/ requires dep, tools/ requires dep) + dep -> wrk --bring -> replace in go-pkgs/go.mod and tools/go.mod
```

## Preconditions

- Git and Go must be available.
- Consumer repo root has NO go.mod; two subdirectories each have go.mod, both requiring dep.
- Consumer cwd is the repo root.

## Steps

1. Create consumer git repo with `go-pkgs/go.mod` and `tools/go.mod`, both requiring `example.com/dep`.
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
	// NO go.mod at root. Two sub-modules both require the dep.
	modDir1 := filepath.Join(consumer, "go-pkgs")
	modDir2 := filepath.Join(consumer, "tools")
	mkdirAll(t, modDir1)
	mkdirAll(t, modDir2)
	writeFile(t, filepath.Join(modDir1, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire "+bringDepModulePath+" v0.0.0\n")
	writeFile(t, filepath.Join(modDir1, "main.go"), "package main\n")
	writeFile(t, filepath.Join(modDir2, "go.mod"), "module example.com/consumer-tools\n\ngo 1.22\n\nrequire "+bringDepModulePath+" v0.0.0\n")
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
	req.ConsumerModDir2 = modDir2
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", depPath}
	return nil
}
```