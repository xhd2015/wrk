# Scenario

**Feature**: wrk --dep reuses an existing live linked worktree of the same dep main under consumer `external/`

```
# Policy A (auto): same as --bring for materialization scope
# ResolveMainRepo(dep) -> live linked WTs under {consumerTop}/external/
# reuse lex-smallest; still apply replace + tidy + /external gitignore
consumer + existing external dep wt -> wrk --dep <dep>
  -> reuse path; replace remains/ensured
  -> stderr reuse warning
```

## Preconditions

- Same dep fixtures as parent (consumer requires dep module).
- Identity key: cleaned absolute dep main path.

## Steps

- Leaves pre-create an external worktree (via first `--dep`), then run a second `--dep`.

## Context

- Stricter than `--bring`: second run must keep replace correct (no soft SKIP path here).
- Stdout: abs path only; stderr: reuse warnings.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureDepReuseHelpersUsed()
	return nil
}

func countDepExternalDirs(t *testing.T, consumerTop string) int {
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

func ensureDepReuseHelpersUsed() {
	_ = countDepExternalDirs
}
```
