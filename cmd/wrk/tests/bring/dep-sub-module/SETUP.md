# Scenario

**Feature**: wrk --bring resolves a dep repo whose Go module lives in a subdirectory (no root go.mod)

```
# consumer requires example.com/dep/sub; dep repo root has no go.mod, sub/ has go.mod
# wrk --bring <dep-root> -> external/mydep worktree + replace => <external>/sub + tidy + gitignore
```

## Preconditions

- Git and Go must be available.
- Consumer cwd must be inside a git work tree with a `go.mod`.
- Dep repo root is a git repo WITHOUT a `go.mod`; the module the consumer requires
  lives in a subdirectory (e.g. `sub/go.mod`), mirroring repositories like `dot-pkgs`
  where the module is `github.com/xhd2015/dot-pkgs/go-pkgs` under `go-pkgs/`.

## Steps

1. Create consumer git repo whose `go.mod` requires `example.com/dep/sub`.
2. Create dep git repo `mydep` with NO root `go.mod`; place `go.mod`
   (`module example.com/dep/sub`) and a package file under `sub/`.
3. Run `wrk --bring <dep-root>` from the consumer.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

const subModulePath = "example.com/dep/sub"

func initBringDepRepoWithSubModule(t *testing.T, workRoot, name string) string {
	t.Helper()
	depPath := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, depPath)
	// Intentionally NO go.mod at the repo root: the module lives in sub/.
	subDir := filepath.Join(depPath, "sub")
	mkdirAll(t, subDir)
	writeFile(t, filepath.Join(subDir, "go.mod"), "module "+subModulePath+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(subDir, "sub.go"), "package sub\n")
	runGitIsolated(t, depPath, "add", ".")
	runGitIsolated(t, depPath, "commit", "-m", "add sub module")
	return depPath
}

func initBringConsumerRepoRequiringSub(t *testing.T, workRoot string) string {
	t.Helper()
	consumer := filepath.Join(workRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	writeFile(t, filepath.Join(consumer, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire "+subModulePath+" v0.0.0\n")
	// Canonicalize so path comparisons match git's toplevel canonicalization
	// (e.g. macOS /var -> /private/var). No-op on filesystems without symlinks.
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	return consumer
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumer := initBringConsumerRepoRequiringSub(t, req.WorkRoot)
	depPath := initBringDepRepoWithSubModule(t, req.WorkRoot, "mydep")

	req.RepoDir = consumer
	req.DepPath = depPath
	req.ConsumerTop = consumer
	req.DepModulePath = subModulePath
	req.Args = []string{"--bring", depPath}
	return nil
}
```
