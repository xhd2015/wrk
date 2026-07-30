# Scenario

**Feature**: wrk --bring matches the correct sub-module when dep has multiple sub-modules

```
# consumer requires example.com/dep/b; dep has a/go.mod (example.com/dep/a) and b/go.mod (example.com/dep/b) -> wrk --bring -> match b/
consumer (requires dep/b) + dep (a/ and b/) -> wrk --bring -> replace => <external>/b
```

## Preconditions

- Git and Go must be available.
- Consumer root go.mod requires `example.com/dep/b`.
- Dep repo root has NO go.mod; `a/go.mod` and `b/go.mod` exist, only `b/` matches.

## Steps

1. Create consumer git repo requiring `example.com/dep/b`.
2. Create dep git repo with sub-modules `a/` and `b/`.
3. Run `wrk --bring <dep>`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

import "path/filepath"

const bringDepModulePathA = "example.com/dep/a"
const bringDepModulePathB = "example.com/dep/b"

func initBringDepRepoMultiSub(t *testing.T, workRoot, name string) string {
	t.Helper()
	depPath := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, depPath)
	for _, sub := range []string{"a", "b"} {
		subDir := filepath.Join(depPath, sub)
		mkdirAll(t, subDir)
		writeFile(t, filepath.Join(subDir, "go.mod"), "module example.com/dep/"+sub+"\n\ngo 1.22\n")
		writeFile(t, filepath.Join(subDir, sub+".go"), "package "+sub+"\n")
	}
	runGitIsolated(t, depPath, "add", ".")
	runGitIsolated(t, depPath, "commit", "-m", "add sub modules")
	return depPath
}

func initBringConsumerRepoRequiringB(t *testing.T, workRoot string) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire "+bringDepModulePathB+" v0.0.0\n")
	writeFile(t, filepath.Join(consumer, "main.go"), "package main\n")
	runGitIsolated(t, consumer, "add", "go.mod", "main.go")
	runGitIsolated(t, consumer, "commit", "-m", "init consumer")
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	return consumer
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepoRequiringB(t, req.WorkRoot)
	depPath := initBringDepRepoMultiSub(t, req.WorkRoot, "mydep")

	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.DepModulePath = bringDepModulePathB
	req.Args = []string{"--bring", depPath}
	return nil
}
```