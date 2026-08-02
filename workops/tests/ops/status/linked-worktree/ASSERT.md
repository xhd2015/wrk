## Expected

- `err` is nil.
- `resp.Status` non-nil.
- `Status.IsWorktree` is true.
- `Status.MainPath` equals seeded MainRepo.
- `Status.CheckoutPath` equals worktree path (or cleans to it).
- `Status.Branch` is non-empty (feature branch name).

## Side Effects

- None (read-only).

## Errors

- None.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp == nil || resp.Status == nil {
		t.Fatal("nil Status report")
	}
	st := resp.Status
	if !st.IsWorktree {
		t.Fatalf("IsWorktree: want true, got false (status=%+v)", st)
	}
	assertPathEqual(t, st.MainPath, req.MainRepo)
	if st.CheckoutPath != "" {
		assertPathEqual(t, st.CheckoutPath, req.WtDir)
	}
	if strings.TrimSpace(st.Branch) == "" {
		t.Fatal("Branch empty")
	}
	// Prefer exact branch when helpers recorded it.
	if req.WtBranch != "" && st.Branch != req.WtBranch {
		t.Fatalf("Branch: got %q want %q", st.Branch, req.WtBranch)
	}
}
```
