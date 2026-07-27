---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; promoted WRK_HOME worktree with task slug.

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertPromotedTaskCreate(t, req, resp, err, taskLikeSpaces)
}
```
