## Expected

- Exit non-zero; stderr task-like / target directory + `-t`/`--task` hint.
- No worktree under WRK_HOME or under WorkRoot named as the long token.

## Exit Code

- non-zero

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertTaskLikeErrorTwoArg(t, resp, err)
	arg := strings.Repeat("a", 121)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, arg))
	// Soft-capped slug path must not appear either (promote must not happen).
	assertFileNotExists(t, wantPromotedWorktree(req, arg))
}
```
