# Scenario

**Feature**: wrk --dep succeeds when the consumer repo has no root go.mod but a subdirectory has one

```
# consumer go.mod lives in go-pkgs/, no go.mod at root -> wrk --dep should scan
# tree for go modules and replace in each matching consumer module
# bug: findGoModDir walks up from cwd to toplevel and fails at root level
# fix: scan all go.mod files under consumer, match against dep module
```

## Preconditions

- Git and Go must be available.
- Consumer repo root has NO go.mod; module `example.com/consumer` lives in `go-pkgs/go.mod`.
- Consumer cwd is the repo root (NOT inside go-pkgs/).

## Steps

1. Create consumer git repo with subdirectory `go-pkgs/` containing `go.mod` that requires `example.com/dep`; no `go.mod` at repo root.
2. Create dep git repo `mydep` with root `go.mod` (`module example.com/dep`).
3. Run `wrk --dep <dep>` from the consumer repo root (not from go-pkgs/).

```go
func Setup(t *testing.T, req *Request) error {
	consumer := filepath.Join(req.WorkRoot, "consumer")
	initGitRepoOnMain(t, consumer)
	// NO go.mod at repo root. Module lives in go-pkgs/.
	modDir := filepath.Join(consumer, "go-pkgs")
	mkdirAll(t, modDir)
	writeFile(t, filepath.Join(modDir, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire "+depModulePath+" v0.0.0\n")
	writeFile(t, filepath.Join(modDir, "main.go"), "package main\n")
	runGitIsolated(t, consumer, "add", ".")
	runGitIsolated(t, consumer, "commit", "-m", "add sub-module")

	dep := initDepRepo(t, req.WorkRoot, "mydep", true)

	// cwd is the consumer repo root (NOT go-pkgs/)
	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	req.RepoDir = consumer
	req.DepPath = dep
	req.ConsumerTop = consumer
	req.ConsumerModDir = modDir
	req.DepModulePath = depModulePath
	req.Args = []string{"--dep", dep}
	return nil
}
```