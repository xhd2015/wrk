## Expected

- Exit non-zero.
- Stderr identifies task-like / not target directory and includes `-t` or `--task` hint
  (must not only be an opaque filesystem ENAMETOOLONG without the task UX).
- No worktree created.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertTaskLikeErrorTwoArg(t, resp, err)
	arg := strings.Repeat("b", 256)
	assertFileNotExists(t, wantPromotedWorktree(req, arg))
}
```
