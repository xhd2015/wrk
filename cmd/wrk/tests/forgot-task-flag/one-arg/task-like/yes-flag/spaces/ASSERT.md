## Expected

- Exit 0; promoted WRK_HOME worktree with task slug.

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
