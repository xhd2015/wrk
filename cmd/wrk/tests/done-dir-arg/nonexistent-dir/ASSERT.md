## Expected

- Exit code non-zero.
- Stderr contains an error indicating the path is not valid.

## Exit Code

- non-zero

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Error should mention the path doesn't exist or isn't a git repo.
	stderr := resp.Stderr
	if !(strings.Contains(stderr, "no such file") ||
		strings.Contains(stderr, "does not exist") ||
		strings.Contains(stderr, "not a git repository") ||
		strings.Contains(stderr, "not a directory")) {
		t.Fatalf("expected error about invalid path, got stderr=%q", stderr)
	}
}
```