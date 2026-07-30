## Expected

- Non-zero exit.
- Stderr includes `Error:`, branch name, both worktree paths, refuse `--merge-back`.
- Both worktrees remain; branch remains; `feature-work` not on main.

## Side Effects

- No merge into main; no worktree removal.

## Errors

- Shared-branch refuse for `--merge-back`.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSharedBranchRefuseError(t, req, resp, "--merge-back")
	assertNoDoneMutations(t, req)
	assertFileExists(t, req.Wt2Dir)
	assertWorktreeListContains(t, req.MainRepo, req.Wt2Dir)
}
```
