# Scenario

**Feature**: wrk --dep matches the correct sub-module when dep has multiple sub-modules

```
# consumer requires example.com/dep/b; dep has a/go.mod (example.com/dep/a) and b/go.mod (example.com/dep/b) -> wrk --dep -> match b/
consumer (requires dep/b) + dep (a/ and b/) -> wrk --dep -> replace => <external>/b
```

## Preconditions

- Git and Go must be available.
- Consumer root go.mod requires `example.com/dep/b`.
- Dep repo root has NO go.mod; `a/go.mod` and `b/go.mod` exist, only `b/` matches.

## Steps

1. Create consumer git repo requiring `example.com/dep/b`.
2. Create dep git repo with sub-modules `a/` and `b/`.
3. Run `wrk --dep <dep>`.

```go
import "path/filepath"

const depModulePathA = "example.com/dep/a"
const depModulePathB = "example.com/dep/b"

func initDepRepoMultiSub(t *testing.T, workRoot, name string) string {
	t.Helper()
	dep := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, dep)
	for _, sub := range []string{"a", "b"} {
		subDir := filepath.Join(dep, sub)
		mkdirAll(t, subDir)
		writeFile(t, filepath.Join(subDir, "go.mod"), "module example.com/dep/"+sub+"\n\ngo 1.22\n")
		writeFile(t, filepath.Join(subDir, sub+".go"), "package "+sub+"\n")
	}
	runGitIsolated(t, dep, "add", ".")
	runGitIsolated(t, dep, "commit", "-m", "add sub modules")
	return dep
}

func initConsumerRepoRequiringB(t *testing.T, workRoot string) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire "+depModulePathB+" v0.0.0\n")
	writeFile(t, filepath.Join(consumer, "main.go"), "package main\n")
	runGitIsolated(t, consumer, "add", "go.mod", "main.go")
	runGitIsolated(t, consumer, "commit", "-m", "init consumer")
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	return consumer
}

func Setup(t *testing.T, req *Request) error {
	consumer := initConsumerRepoRequiringB(t, req.WorkRoot)
	dep := initDepRepoMultiSub(t, req.WorkRoot, "mydep")

	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.DepModulePath = depModulePathB
	req.Args = []string{"--dep", dep}
	return nil
}
```