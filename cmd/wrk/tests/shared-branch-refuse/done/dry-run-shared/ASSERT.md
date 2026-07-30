## Expected

- Non-zero exit (fail closed — dry-run must not exit 0 with a plan).
- Stderr includes `Error:`, branch, both paths, refuse `--done`.
- Worktrees and branch unchanged; `feature-work` not on main.

## Side Effects

- None (refuse before plan apply; no successful dry-run plan exit).

## Errors

- Shared-branch refuse under `--dry-run`.

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
