## Expected

- Exit 0 with a newline-terminated expanded plan for the dependency peel.
- Peel line uses **relative display path** of the linked dep checkout vs app cwd:
  `would: peel external/dep` (not bare `would: peel dep` alone as full path).
- The plan orders generated commit, merge-back, tag/pin (cascade or under-peel),
  then **ship tail** push + sync (once, after peels), then reinstall.
- No ref or worktree mutation occurs.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertExit0(t, resp)
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("stdout lacks final newline: %q", resp.Stdout)
	}
	// Display path: nested dep under app/external/dep with RepoDir = app main.
	wantPeel := "would: peel external/dep"
	if !strings.Contains(resp.Stdout, wantPeel) {
		t.Fatalf("want peel display %q\nstdout:\n%s", wantPeel, resp.Stdout)
	}
	// Bare basename alone must not be the only peel form for nested checkout.
	if strings.Contains(resp.Stdout, "would: peel dep\n") && !strings.Contains(resp.Stdout, wantPeel) {
		t.Fatalf("nested peel must not be bare basename only; stdout:\n%s", resp.Stdout)
	}
	// Ship tail: push/sync after peel land plan (not under-peel); reinstall last.
	assertContainsInOrder(t, resp.Stdout,
		wantPeel,
		"generate",
		"commit",
		"merge",
		"tag",
		"push",
		"sync",
		"reinstall",
	)
	// Pin may be cascade "would: pin" and/or under-peel legacy pin — require present.
	if !strings.Contains(resp.Stdout, "pin") {
		t.Fatalf("want pin in dry-run plan\nstdout:\n%s", resp.Stdout)
	}
	// Sync is ship-tail only (not repeated under each peel before tag/pin).
	if strings.Count(resp.Stdout, "would: sync linked worktrees") < 1 {
		t.Fatalf("want ship-tail would: sync linked worktrees\nstdout:\n%s", resp.Stdout)
	}
	if got := git(t, req.DepMain, "rev-parse", "HEAD"); got != req.BeforeDep {
		t.Fatalf("dry-run mutated dep: %s != %s", got, req.BeforeDep)
	}
	if got := git(t, req.MainRepo, "rev-parse", "HEAD"); got != req.BeforeMain {
		t.Fatalf("dry-run mutated main: %s != %s", got, req.BeforeMain)
	}
}
```
