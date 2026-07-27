
## Expected

- Non-zero exit.
- Follow-up file empty.
- Fake interactive shell not launched.
- Stderr mentions not a git repository (or similar create failure).

## Exit Code

- Non-zero

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertFollowupEmpty(t, resp)
	assertFakeShellNotLaunched(t, req)
	assert.Output(t, resp.Stderr, `<contains>
not a git repository
</contains>`)
}
```
