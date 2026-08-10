## Expected Output

```
==== unwind graph (module) ====
  dir  latest    status
  .    v0.0.1    owned-changed  next=v0.0.2
…
```

(Field labels implementer-owned; latest + next tag names locked. Dir identity, not full path.)

## Expected

- Exit code 0.
- Human banners present.
- Module dir `.` listed.
- Output mentions latest tag `v0.0.1` and next `v0.0.2` (or equivalent
  owned-changed / next-tag language that includes both tag names).
- Does **not** create new tags (zero mutations; next tag remains absent on repo).
- Zero HEAD mutations.

## Side Effects

- None (read-only; no tag-next apply).

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertShowGraphHumanBanners(t, resp.Stdout)
	assertModuleDirListed(t, resp.Stdout, ".")
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertHumanNoFlatFullPathEdges(t, resp.Stdout)
	out := resp.Stdout
	latest := req.OldRequireVersion
	next := req.ExpectedPinVersion
	if latest == "" {
		latest = unwindApplyOldTag
	}
	if next == "" {
		next = unwindApplyNextTag
	}
	if !strings.Contains(out, latest) {
		t.Fatalf("module status must mention latest tag %q; stdout:\n%s", latest, out)
	}
	if !strings.Contains(out, next) {
		lower := strings.ToLower(out)
		if !strings.Contains(lower, "owned-changed") && !strings.Contains(lower, "owned_changed") &&
			!strings.Contains(lower, "next") {
			t.Fatalf("module status must mention next tag %q or owned-changed; stdout:\n%s", next, out)
		}
		if !strings.Contains(out, next) {
			t.Fatalf("module status must include planned next tag %q; stdout:\n%s", next, out)
		}
	}
	// Read-only: next tag must not have been created on the repo.
	if tagRefExists(t, req.MainRepo, next) {
		t.Fatalf("show-graph must not create next tag %q", next)
	}
	assertShowGraphZeroMutations(t, req)
}
```
