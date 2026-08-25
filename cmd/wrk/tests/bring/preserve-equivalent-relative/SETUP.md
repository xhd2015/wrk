# Scenario

**Feature**: `--bring` does not rewrite nested intra-repo relative replaces inside the brought checkout

```
# consumer requires example.com/dep
# dep has root go.mod + cmd/ with replace => ../
cwd=consumer -> wrk --bring <dep>
  -> external worktree; consumer abs replace
  -> external/.../cmd/go.mod keeps => ../
```

## Steps

1. Seed consumer requiring `example.com/dep`.
2. Seed dep with root module + nested `cmd` relative replace.
3. Run `--bring` from consumer.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

const bringDepCmdModulePath = "example.com/dep/cmd"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepo(t, req.WorkRoot, true)

	dep := filepath.Join(req.WorkRoot, "mydep")
	initGitRepoOnMain(t, dep)
	writeFile(t, filepath.Join(dep, "go.mod"), "module "+bringDepModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(dep, "dep.go"), "package dep\n")
	cmdDir := filepath.Join(dep, "cmd")
	mkdirAll(t, cmdDir)
	writeFile(t, filepath.Join(cmdDir, "go.mod"),
		"module "+bringDepCmdModulePath+"\n\ngo 1.22\n\nrequire "+bringDepModulePath+" v0.0.0\n\nreplace "+bringDepModulePath+" => ../\n")
	writeFile(t, filepath.Join(cmdDir, "main.go"), "package main\n")
	runGitIsolated(t, dep, "add", "-A")
	runGitIsolated(t, dep, "commit", "-m", "dep + cmd")

	dep, err := filepath.EvalSymlinks(dep)
	if err != nil {
		t.Fatalf("eval symlinks dep: %v", err)
	}
	consumer, err = filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks consumer: %v", err)
	}

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.ConsumerModDir = consumer
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", dep}
	return nil
}
```
