## Expected

- Exit 0; promoted WRK_HOME worktree for the long token.
- No spawn under WorkRoot named as the long token.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	arg := strings.Repeat("a", 121)
	assertPromotedTaskCreate(t, req, resp, err, arg)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, arg))
}
```
