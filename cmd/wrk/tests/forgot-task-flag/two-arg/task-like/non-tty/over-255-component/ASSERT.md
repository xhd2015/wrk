## Expected

- Exit 0; promoted WRK_HOME worktree (slug fitted to name budget).
- Path last component and branch ≤ 255 bytes.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	arg := strings.Repeat("b", 256)
	assertPromotedTaskCreate(t, req, resp, err, arg)
	got := strings.TrimSpace(resp.Stdout)
	if len(filepath.Base(got)) > 255 {
		t.Fatalf("path base exceeds 255 bytes: %q len=%d", filepath.Base(got), len(filepath.Base(got)))
	}
	br := wantPromotedBranch(arg)
	if len(br) > 255 {
		t.Fatalf("branch exceeds 255 bytes: %q len=%d", br, len(br))
	}
}
```
