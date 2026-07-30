# Scenario

**Feature**: wrk --bring --no-dep materializes external worktree only (skip replace + tidy)

```
# --no-dep with --bring: still create/reuse external wt + /external gitignore
# skip module match scan, replace.ReplaceIn, and go mod tidy entirely
consumer (git) + dep path -> wrk --bring <dep> --no-dep
  -> external worktree under consumer/external/
  -> go.mod byte-identical (no replace)
  -> no go mod tidy
  -> stdout: <external-abs>\n
```

## Preconditions

- Git and Go available (same as parent `bring/`).
- `--no-dep` is long-only and valid only with `--bring`.

## Steps

- Leaves build consumer + dep fixtures via `initBringConsumerRepo` / `initBringDepRepo`.
- `req.Args` include `--bring`, dep path, and `--no-dep` (plus optional `-v`).

## Context

- Fast path: no SKIP messages required (module analyse skipped).
- With `-v`: may log/stream git worktree add; must never log `go … mod tidy`.

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
	ensureBringNoDepHelpersUsed()
	return nil
}

func ensureBringNoDepHelpersUsed() {
	_ = snapshotBringGoMod
	_ = assertBringGoModUnchanged
}

// snapshotBringGoMod writes consumer go.mod bytes to WorkRoot for later compare.
func snapshotBringGoMod(t *testing.T, req *Request, modDir string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(modDir, "go.mod"))
	if err != nil {
		t.Fatalf("snapshot go.mod: %v", err)
	}
	path := filepath.Join(req.WorkRoot, "go.mod.before")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write go.mod.before: %v", err)
	}
}

// assertBringGoModUnchanged fails if consumer go.mod differs from snapshotBringGoMod.
func assertBringGoModUnchanged(t *testing.T, req *Request, modDir string) {
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
		t.Fatalf("go.mod changed under --no-dep\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
```
