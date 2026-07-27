## Expected

- Non-zero exit.
- Flag is accepted (not "unrecognized flag: --new-window").
- Stderr clearly indicates unsupported platform / macOS-only space.
- Worktree may or may not exist (create may run before window); do not require cleanup.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero on non-darwin window, stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected platform error after --new-window is implemented; got %q", resp.Stderr)
	}
	s := strings.ToLower(resp.Stderr)
	if strings.Contains(s, "unsupported") || strings.Contains(s, "macos") || strings.Contains(s, "darwin") {
		return
	}
	assert.Output(t, resp.Stderr, `<contains>
unsupported
</contains>`)
}
```
