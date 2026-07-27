## Expected

- Exit non-zero; stderr task-like / not source directory + `-t`/`--task` hint.
- No worktree created under WRK_HOME for this task.

## Exit Code

- non-zero

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertTaskLikeErrorOneArg(t, resp, err)
	assertFileNotExists(t, wantPromotedWorktree(req, taskLikeSpaces))
}
```
