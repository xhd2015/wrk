# Scenario

**Feature**: wrk --dep --no-dep keeps strict pre-checks; worktree only (no replace/tidy)

```
# --dep --no-dep: still strict analyse first (must be go module + dependency)
# then external worktree + gitignore; skip replace + go mod tidy
consumer requires dep -> wrk --dep <dep> --no-dep
  -> external wt; go.mod unchanged
# not a dependency -> hard error; no external wt (analyse before worktree)
```

## Preconditions

- Same as parent `dep/` (git + go).
- Strict path: not-a-dependency / not-go-module still fail before worktree create.

## Steps

- Leaves use `initConsumerRepo` / `initDepRepo`.
- Snapshot go.mod when asserting byte-identity.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	ensureDepNoDepHelpersUsed()
	return nil
}

func ensureDepNoDepHelpersUsed() {
	_ = snapshotDepGoMod
	_ = assertDepGoModUnchanged
}

func snapshotDepGoMod(t *testing.T, req *Request, modDir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		t.Fatalf("snapshot go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(req.WorkRoot, "go.mod.before"), data, 0o644); err != nil {
		t.Fatalf("write go.mod.before: %v", err)
	}
}

func assertDepGoModUnchanged(t *testing.T, req *Request, modDir string) {
	t.Helper()
	before, err := os.ReadFile(filepath.Join(req.WorkRoot, "go.mod.before"))
	if err != nil {
		t.Fatalf("read go.mod.before: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("go.mod changed under --dep --no-dep\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
```
