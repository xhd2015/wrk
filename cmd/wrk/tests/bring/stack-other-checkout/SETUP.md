# Scenario

**Feature**: `--bring` fans absolute replace across the unwind stack (not only invoke checkout)

```
# primary requires dep + kool; replace kool => ./external/kool (own git)
# kool already requires dep
cwd=primary -> wrk --bring <dep>
  -> external worktree under primary/external/
  -> replace on example.com/app (checkout .)
  -> replace on example.com/kool (checkout external/kool)
  -> stdout: external abs path only (silent replace/tidy)
```

## Steps

1. Seed primary + independent `external/kool` git repo + filesystem replace.
2. Seed dep git repo with module `example.com/dep`.
3. Run `--bring` from primary.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true

	primary := filepath.Join(req.WorkRoot, "app")
	initGitRepoOnMain(t, primary)
	writeFile(t, filepath.Join(primary, "go.mod"), "module example.com/app\n\ngo 1.22\n")

	kool := filepath.Join(primary, "external", "kool")
	initGitRepoOnMain(t, kool)
	writeFile(t, filepath.Join(kool, "go.mod"),
		"module example.com/kool\n\ngo 1.22\n\nrequire "+bringDepModulePath+" v0.0.0\n")
	writeFile(t, filepath.Join(kool, "kool.go"), "package kool\n")
	runGitIsolated(t, kool, "add", "-A")
	runGitIsolated(t, kool, "commit", "-m", "kool requirer")

	runBringGo(t, primary, "mod", "edit", "-require="+bringDepModulePath+"@v0.0.0")
	runBringGo(t, primary, "mod", "edit", "-require=example.com/kool@v0.0.0")
	runBringGo(t, primary, "mod", "edit", "-replace=example.com/kool=./external/kool")
	runGitIsolated(t, primary, "add", "-A")
	runGitIsolated(t, primary, "commit", "-m", "primary + replace kool")

	dep := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	primary, err := filepath.EvalSymlinks(primary)
	if err != nil {
		t.Fatalf("eval symlinks primary: %v", err)
	}
	kool, err = filepath.EvalSymlinks(kool)
	if err != nil {
		t.Fatalf("eval symlinks kool: %v", err)
	}

	req.RepoDir = primary
	req.ConsumerTop = primary
	req.ConsumerModDir = primary
	req.ConsumerModDir2 = kool
	req.DepPath = dep
	req.Args = []string{"--bring", dep}
	return nil
}
```
