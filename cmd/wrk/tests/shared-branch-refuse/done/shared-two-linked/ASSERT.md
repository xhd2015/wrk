## Expected

- Non-zero exit.
- Stderr includes `Error:`, the shared branch name, both worktree paths (or basenames),
  and refuse language naming `--done`.
- Primary worktree directory still exists and remains in `git worktree list`.
- Second worktree still exists and remains listed.
- Branch still exists on main; `feature-work` not on main (no merge).

## Side Effects

- No worktree remove; no branch `-D`; no merge into main.

## Errors

- Shared-branch refuse for `--done`.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSharedBranchRefuseError(t, req, resp, "--done")
	assertNoDoneMutations(t, req)
	assertFileExists(t, req.Wt2Dir)
	assertWorktreeListContains(t, req.MainRepo, req.Wt2Dir)
}
```
