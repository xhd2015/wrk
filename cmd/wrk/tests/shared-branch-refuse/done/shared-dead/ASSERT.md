## Expected

- Non-zero exit.
- Stderr includes `Error:`, branch name, both worktree paths (or basenames),
  refuse `--done`, and a dead-worktree prune hint:
  `worktree prune` via `git -C <main> …`.
- Primary worktree still exists; branch kept; no merge to main.

## Side Effects

- No worktree remove on primary; no prune performed by wrk itself.

## Errors

- Shared-branch refuse including dead registration + prune command.

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
	assertDeadPruneHint(t, req, resp)
	assertNoDoneMutations(t, req)
}
```
