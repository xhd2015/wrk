# Scenario

**Feature**: wrk --bring reuses an existing live linked worktree of the same dep main under consumer `external/`

```
# Policy A (auto): ResolveMainRepo(dep) -> live linked WTs under {consumerTop}/external/
# if any -> reuse lex-smallest abs path; no new worktree/branch; still gitignore + optional replace
# multi -> reuse smallest + stderr count + also-present warnings
# stdout: abs path only; stderr: reuse warnings
consumer + dep already under external/ -> wrk --bring <dep>
  -> reuse existing path (no -1)
  -> stderr: wrk: warning: <basename> already exists under external/; reusing <absPath>
```

## Preconditions

- Same bring fixtures as parent: consumer + dep under `WorkRoot`.
- Identity key: cleaned absolute `ResolveMainRepo(dep)` path; live = path exists + linked worktree.
- Scope is **only** `{consumerTop}/external/` (not other linked WTs of the dep).

## Steps

- Leaves first materialize one or more external worktrees of the dep under consumer `external/`.
- The doctest `Run` is the subsequent `--bring` that must reuse (or create for a different dep).

## Context

- Suggested stderr (single): `wrk: warning: <basename> already exists under external/; reusing <absPath>`
- Suggested stderr (multi): `already has <N> worktrees under external/; reusing …` plus `also present: <otherAbsPath>`
- No `--force` / override flags.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Parent bring Setup already skipIfNoGit + go check.
	ensureBringReuseHelpersUsed()
	return nil
}

// countExternalDirs returns the number of direct child entries under consumer/external.
func countExternalDirs(t *testing.T, consumerTop string) int {
	t.Helper()
	ext := filepath.Join(consumerTop, "external")
	entries, err := os.ReadDir(ext)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("readdir %s: %v", ext, err)
	}
	return len(entries)
}

func ensureBringReuseHelpersUsed() {
	_ = countExternalDirs
}
```
