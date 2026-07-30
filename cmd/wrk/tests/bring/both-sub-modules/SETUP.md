# Scenario

**Feature**: wrk --bring when both consumer and dep go.mod live in subdirectories

```
# consumer go.mod in go-pkgs/ requires example.com/dep/sub; dep go.mod in sub/ -> wrk --bring -> replace in go-pkgs/go.mod => <external>/sub
consumer (go-pkgs/ requires dep/sub) + dep (sub/ has go.mod) -> wrk --bring -> external worktree + replace
```

## Preconditions

- Git and Go must be available.
- Consumer repo root has NO go.mod; `go-pkgs/go.mod` requires `example.com/dep/sub`.
- Dep repo root has NO go.mod; `sub/go.mod` has module `example.com/dep/sub`.
- Consumer cwd is the repo root.

## Steps

1. Create consumer git repo with `go-pkgs/go.mod` requiring `example.com/dep/sub`.
2. Create dep git repo with `sub/go.mod` (`module example.com/dep/sub`), no root `go.mod`.
3. Run `wrk --bring <dep>` from the consumer repo root.

```go
import (
	"github.com/xhd2015/doctest/session"
)

import "path/filepath"

const depSubModulePath = "example.com/dep/sub"

func initBringDepRepoWithSubBoth(t *testing.T, workRoot, name string) string {
	t.Helper()
	depPath := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, depPath)
	subDir := filepath.Join(depPath, "sub")
	mkdirAll(t, subDir)
	writeFile(t, filepath.Join(subDir, "go.mod"), "module "+depSubModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(subDir, "sub.go"), "package sub\n")
	runGitIsolated(t, depPath, "add", ".")
	runGitIsolated(t, depPath, "commit", "-m", "add sub module")
	return depPath
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	modDir := filepath.Join(consumer, "go-pkgs")
	mkdirAll(t, modDir)
	writeFile(t, filepath.Join(modDir, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire "+depSubModulePath+" v0.0.0\n")
	writeFile(t, filepath.Join(modDir, "main.go"), "package main\n")
	runGitIsolated(t, consumer, "add", ".")
	runGitIsolated(t, consumer, "commit", "-m", "add sub-module")

	depPath := initBringDepRepoWithSubBoth(t, req.WorkRoot, "mydep")

	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.ConsumerModDir = modDir
	req.DepModulePath = depSubModulePath
	req.Args = []string{"--bring", depPath}
	return nil
}
```