## Expected

- Exit non-zero; stderr task-like / not source directory + `-t`/`--task` hint.
- No worktree created under WRK_HOME for this task.

## Exit Code

- non-zero

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertTaskLikeErrorOneArg(t, resp, err)
	assertFileNotExists(t, wantPromotedWorktree(req, taskLikeSpaces))
}
```
