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
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertTaskLikeErrorTwoArg(t, resp, err)
	arg := strings.Repeat("b", 256)
	assertFileNotExists(t, wantPromotedWorktree(req, arg))
}
```
