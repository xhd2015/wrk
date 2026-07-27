## Expected

- Non-zero exit code.
- Stderr explains that home / default root is unusable (missing or not a directory).
- Stderr must **not** require or advertise `Projects` / `~/Projects` as the default root.
- Stdout is empty.

## Errors

- Empty, unresolvable, or non-directory `$HOME` → clear error mentioning home or `~` (not Projects).

## Exit Code

- Non-zero

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero when HOME is not a directory, got 0 stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty on home error, got %q", resp.Stdout)
	}
	// Product must not still tell users the default is ~/Projects.
	if strings.Contains(resp.Stderr, "Projects") {
		t.Fatalf("stderr must not require Projects as default root; got %q", resp.Stderr)
	}
	// Soft: message should orient on home / ~ / not-a-directory.
	lower := strings.ToLower(resp.Stderr)
	if !strings.Contains(lower, "home") &&
		!strings.Contains(resp.Stderr, "~") &&
		!strings.Contains(lower, "not a directory") {
		t.Fatalf("stderr should mention home, ~, or not-a-directory; got %q", resp.Stderr)
	}
}
```
