## Expected

- Exit 0; stdout is WRK_HOME-managed worktree path with task slug.
- Branch includes slug; prose path under WorkRoot is not used as fixed target-dir.
- Stderr may include warning / Treat as --task prompt text.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPromotedTaskCreate(t, req, resp, err, taskLikeSpaces)
	// Optional: warning tokens when present
	low := strings.ToLower(resp.Stderr)
	if low != "" && !strings.Contains(low, "task") && !strings.Contains(low, "treat") {
		// Allow empty stderr if implementer auto-accepts without re-echoing; prefer task mention when any stderr.
		_ = low
	}
}
```
