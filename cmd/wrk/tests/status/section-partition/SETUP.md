# Scenario

**Feature**: pure helpers partition status paths into ordered primary vs external

```
# main + ListLinked porcelain → primary (main first, then linked order)
# scan paths not primary → external (lexicographic by normalized path)
mainRoot + scanPaths + linkedOrdered
  -> wrkcli.PartitionStatusPaths
  -> StatusPathLists{Primary, External}
```

## Preconditions

- Package `github.com/xhd2015/wrk/wrkcli` is importable from this module.
- Nested root: no inheritance from `cmd/wrk/tests` or `status/` (own `DOCTEST.md`).
- Pure helper only — no git, no wrk binary, no `WRK_HOME`, no filesystem mutations.
- Fixture paths are **synthetic absolute** strings under `/section-partition-fixture`
  (need not exist on disk). Product normalizes via `storage.NormalizePath`;
  Assert normalizes both actual and expected the same way.

## Steps

1. Leaves assign `MainRoot`, `ScanPaths`, `LinkedOrdered`.
2. Leaves assign `WantPrimary` / `WantExternal` with the same raw fixture paths.
3. Root `Run` calls `wrkcli.PartitionStatusPaths` and returns the lists.

## Context

- **Fixture vocabulary** (helpers below):
  - `Main` — main checkout
  - `LinkedInTree` — in-tree linked worktree under main
  - `LinkedLate` / `LinkedEarly` — out-of-tree WRK WTs; **Late** sorts after
    **Early** by path, so ListLinked order `[Late, Early]` ≠ path sort
  - `DeadLinked` — prunable/dead path present only in linked list
  - `NestedTaskHub` / `NestedExternalChild` / `NestedToolsChild` — scan-only
    nested/dep repos (external)
- Primary never includes nested/dep paths; external never includes main or
  ListLinked members.
- Dedup is by normalized path.

```go
import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xhd2015/wrk/wrkcli/storage"
)

// Synthetic absolute fixture root (paths need not exist on disk).
const fixtureRoot = "/section-partition-fixture"

func Setup(t *testing.T, req *Request) error {
	// Root setup is intentionally light: pure-helper tree has no session
	// binary build and no shared mutable fixtures. Leaves fill Request fields.
	if req.ScanPaths == nil {
		req.ScanPaths = []string{}
	}
	if req.LinkedOrdered == nil {
		req.LinkedOrdered = []string{}
	}
	return nil
}

// fix joins absolute synthetic fixture segments under fixtureRoot.
func fix(parts ...string) string {
	all := append([]string{fixtureRoot}, parts...)
	return filepath.Clean(filepath.Join(all...))
}

func pathMain() string            { return fix("repo", "myrepo") }
func pathLinkedInTree() string    { return fix("repo", "myrepo", "wt-linked") }
func pathLinkedLate() string      { return fix("wrk-home", "worktrees", "zzz-late") }
func pathLinkedEarly() string     { return fix("wrk-home", "worktrees", "aaa-early") }
func pathDeadLinked() string      { return fix("wrk-home", "worktrees", "dead-prunable") }
func pathNestedTaskHub() string   { return fix("repo", "myrepo", "task-hub") }
func pathNestedExternal() string  { return fix("repo", "myrepo", "external", "child") }
func pathNestedTools() string     { return fix("repo", "myrepo", "tools", "child") }

func normPath(p string) string {
	return storage.NormalizePath(p)
}

func normPaths(ps []string) []string {
	if ps == nil {
		return []string{}
	}
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = normPath(p)
	}
	return out
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertPathListsEqual(t *testing.T, got, want []string, label string) {
	t.Helper()
	g := normPaths(got)
	w := normPaths(want)
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("%s mismatch\n got: %#v\nwant: %#v", label, g, w)
	}
}

func assertPartition(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertPathListsEqual(t, resp.Primary, req.WantPrimary, "Primary")
	assertPathListsEqual(t, resp.External, req.WantExternal, "External")
}
```
