# Scenario

**Feature**: multi-bring preflight with multiple ambiguous basenames (one-by-one Select + will bring plan)

```
# two+ --bring args; each ambiguous basename Select once, left→right
# after all resolves succeed -> stderr "will bring:" plan (multi only)
# duplicate resolved abs after selection -> hard error; no plan; no create
# stdout remains only external abs paths (one per line) on success
consumer + ambiguous mydep + ambiguous otherlib
  -> wrk --bring mydep --bring otherlib
  -> Select per arg -> will bring plan -> apply
```

## Preconditions

- Parent `multi/` helpers (`initMultiBringConsumerWithTwoRequires`, `initMultiBringDepRepo`,
  `multiRecordSavedProject`, …) and root bring helpers are available.
- L2 only: leaves set `req.InProcess = true`, `req.BasenameEnv = "WRK_BASENAME_CONFIRM=1"`,
  and multi-line `req.StdinInput` for Select indices.

## Steps

- Leaves seed four saved dep repos under distinct parents so each basename is ambiguous:
  - `aaa/mydep` + `zzz/mydep` → module `example.com/dep1`
  - `aaa/otherlib` + `zzz/otherlib` → module `example.com/dep2`
- Consumer requires both modules; no local cwd copies of the basenames.
- Candidates are listed in lexicographic abs-path order (same as single-arg basename fallback).

## Context

- **will bring:** printed on **stderr** only when `len(--bring) > 1`, after preflight
  resolves all args successfully, **before** any external path on stdout / any worktree create.
- One plan line per arg: display key = raw bring arg (basename or path) + `→` + resolved abs path.
- Single-arg bring must **not** print `will bring:` (noise reduction; covered under basename-fallback).
- Duplicate detection after resolve still uses `wrk: duplicate --bring path: <abs>` (no plan, no create).

```go
import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	ensureMultiPreflightAmbiguousHelpersUsed()
	return nil
}

// multiPreflightAbs canonicalizes a path for plan/duplicate assertions.
func multiPreflightAbs(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
}

// multiPreflightSorted returns lexicographically sorted absolute paths.
func multiPreflightSorted(t *testing.T, paths ...string) []string {
	t.Helper()
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, multiPreflightAbs(t, p))
	}
	sort.Strings(out)
	return out
}

// multiPreflightRecordSaved registers a main repo via wrk --add (projects.json).
func multiPreflightRecordSaved(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	multiRecordSavedProject(t, req, repoPath)
}

// multiPreflightPlanLineHas reports whether stderr has a will-bring plan line for
// key → abs (allows flexible whitespace; accepts → or ->).
func multiPreflightPlanLineHas(stderr, key, abs string) bool {
	for _, line := range strings.Split(stderr, "\n") {
		if !strings.Contains(line, key) || !strings.Contains(line, abs) {
			continue
		}
		if strings.Contains(line, "→") || strings.Contains(line, "->") {
			return true
		}
	}
	return false
}

func ensureMultiPreflightAmbiguousHelpersUsed() {
	_ = multiPreflightAbs
	_ = multiPreflightSorted
	_ = multiPreflightRecordSaved
	_ = multiPreflightPlanLineHas
}
```
