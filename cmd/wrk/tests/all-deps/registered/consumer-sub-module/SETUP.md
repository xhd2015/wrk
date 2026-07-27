# Scenario

**Feature**: wrk --all-deps works when the consumer module lives in a subdirectory

```
# consumer go.mod in go-pkgs/ (no root go.mod); registered deps -> scan + link both
projects.json (mydep1, mydep2) + consumer go-pkgs/ requires dep1+dep2 -> wrked 2 deps
```

## Steps

1. Create and register `mydep1` and `mydep2`.
2. Create a consumer git repo with `go-pkgs/go.mod` requiring both deps; no root `go.mod`.
3. Run `wrk --all-deps` from the consumer repo root.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	dep2 := allDepsDepDir(req.WorkRoot, "mydep2")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	initAllDepsRepo(t, dep2, "example.com/dep2", "dep2")
	registerAllDepsProjects(t, req, dep1, dep2)

	consumer := filepath.Join(req.WorkRoot, "consumer")
	mkdirAll(t, consumer)
	runGitIsolated(t, consumer, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, consumer, "config", "user.email", "test@test.com")
	runGitIsolated(t, consumer, "config", "user.name", "Test")
	modDir := filepath.Join(consumer, "go-pkgs")
	mkdirAll(t, modDir)
	writeFile(t, filepath.Join(modDir, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire example.com/dep1 v0.0.0\nrequire example.com/dep2 v0.0.0\n")
	writeFile(t, filepath.Join(modDir, "main.go"), "package main\n")
	runGitIsolated(t, consumer, "add", ".")
	runGitIsolated(t, consumer, "commit", "-m", "init consumer")
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps"}
	return nil
}
```