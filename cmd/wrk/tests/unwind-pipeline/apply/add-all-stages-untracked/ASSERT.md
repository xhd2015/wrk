## Expected

- Exit 0 after linked peel with gen-commit + `--add-all`.
- Dependency main advances; `change.txt` is present in dep main **HEAD** tree
  (untracked was staged via `--add-all` and committed).
- Linked worktree remains after merge-back.

## Side Effects

- Gen-commit pre-step honors `--add-all` (`git add -A` before generate).

## Exit Code

- 0

```go
import (
	"os"

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
	if !pathInHEAD(t, req.DepMain, req.UntrackedName) {
		t.Fatalf("with --add-all, %q must be in dep main HEAD after gen-commit land", req.UntrackedName)
	}
	if _, e := os.Stat(req.DepWorktree); e != nil {
		t.Fatalf("linked worktree must remain after merge-back: %v", e)
	}
}
```
