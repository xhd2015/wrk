## Expected

- Exit 0 after the linked worktree is committed and landed.
- Observable progress orders commit → merge-back → push/sync; the dependency
  main advances and its linked worktree remains present.
- No module cascade tags required for this peel-only fixture.

## Side Effects

- The dependency main advances; post-land stages run only after merge-back.

## Exit Code

- 0

```go
import (
	"os"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertExit0(t, resp)
	combined := resp.Stdout + "\n" + resp.Stderr
	assertContainsInOrder(t, combined, "commit", "merge-back", "push")
	if !strings.Contains(combined, "sync") {
		t.Fatalf("want sync in progress output\n%s", combined)
	}
	if got := git(t, req.DepMain, "rev-parse", "HEAD"); got == req.BeforeDep {
		t.Fatal("linked peel did not advance dependency main")
	}
	if _, e := os.Stat(req.DepWorktree); e != nil {
		t.Fatalf("linked worktree must remain after merge-back: %v", e)
	}
}
```
