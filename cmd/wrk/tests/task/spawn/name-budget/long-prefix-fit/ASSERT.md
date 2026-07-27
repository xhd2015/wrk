
## Expected

- Exit 0; worktree created under WRK_HOME.
- `filepath.Base(path)` and branch each ≤ 255 bytes.
- `Base == basename + "-" + branch`.
- Fitted slug shorter than soft-cap 64-rune slug (prefix+64 would exceed 255).

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	basename := longBasename(longRepoBasenameLen)
	assertNameBudgetOK(t, req, resp, err, basename, req.TaskDesc)
}
```
