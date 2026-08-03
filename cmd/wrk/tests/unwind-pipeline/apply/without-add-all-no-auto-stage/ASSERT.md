## Expected

- Exit 0 when staged content is enough for gen-commit + land.
- Dependency main advances and **includes** `tracked.txt` in HEAD.
- Dependency main HEAD does **not** include `leftover.txt` (untracked was not
  auto-staged by an unconditional `git add -A`).
- If the linked worktree still exists, `leftover.txt` remains untracked there
  (or at least is not silently forced into the land commit).

## Side Effects

- Apply gen-commit pre-step must **not** always `git add -A` when `--add-all` is absent.
- Today production always stages before gen-commit → **RED** until fixed.

## Exit Code

- 0

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertExit0(t, resp)
	if got := git(t, req.DepMain, "rev-parse", "HEAD"); got == req.BeforeDep {
		t.Fatal("linked peel did not advance dependency main")
	}
	if !pathInHEAD(t, req.DepMain, req.TrackedName) {
		t.Fatalf("staged %q must be in dep main HEAD", req.TrackedName)
	}
	if pathInHEAD(t, req.DepMain, req.UntrackedName) {
		t.Fatalf("without --add-all, untracked %q must not be forced into the landed commit (unconditional git add -A is wrong)", req.UntrackedName)
	}
	// If worktree kept and leftover file still exists, it should remain untracked
	// (not silently staged without --add-all).
	if _, e := os.Stat(req.DepWorktree); e == nil {
		p := filepath.Join(req.DepWorktree, req.UntrackedName)
		if _, err := os.Stat(p); err == nil {
			if !isUntracked(t, req.DepWorktree, req.UntrackedName) {
				t.Fatalf("without --add-all, %q should remain untracked on worktree if still present; porcelain=%q",
					req.UntrackedName, porcelainFor(t, req.DepWorktree, req.UntrackedName))
			}
		}
	}
}
```
