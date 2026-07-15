## Expected

- Exit code non-zero.
- Stderr mentions task description / not a target directory and includes `-t` or `--task` hint.
- No worktree created at prose path or under WRK_HOME for this task.

## Exit Code

- non-zero

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertTaskLikeErrorTwoArg(t, resp, err)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, taskLikeSpaces))
	assertFileNotExists(t, wantPromotedWorktree(req, taskLikeSpaces))
}
```
