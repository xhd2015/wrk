## Expected

- Exit 0; promoted WRK_HOME worktree with task slug.
- No fixed-path spawn at `{WorkRoot}/fix the login bug`.

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPromotedTaskCreate(t, req, resp, err, taskLikeSpaces)
}
```
