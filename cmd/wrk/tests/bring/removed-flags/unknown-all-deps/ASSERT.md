## Expected

- Non-zero exit code.
- Stderr indicates unknown/invalid flag and names `--all-deps`.
- No `external/` directory created under the consumer.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for removed --all-deps, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "--all-deps")
	ls := strings.ToLower(resp.Stderr)
	ok := strings.Contains(ls, "unknown") ||
		strings.Contains(ls, "unexpected") ||
		strings.Contains(ls, "invalid") ||
		strings.Contains(ls, "not defined") ||
		strings.Contains(ls, "flag provided")
	if !ok {
		t.Fatalf("stderr should indicate unknown/invalid flag for --all-deps; got %q", resp.Stderr)
	}
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
}
```
