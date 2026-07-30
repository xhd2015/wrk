## Expected

- Non-zero exit code (flag hard-removed; not a successful external worktree).
- Stderr indicates unknown/invalid flag and names `--dep`.
- No `external/` directory under the consumer (mode never ran).

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
		t.Fatalf("expected non-zero exit for removed --dep, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "--dep")
	ls := strings.ToLower(resp.Stderr)
	ok := strings.Contains(ls, "unknown") ||
		strings.Contains(ls, "unexpected") ||
		strings.Contains(ls, "invalid") ||
		strings.Contains(ls, "not defined") ||
		strings.Contains(ls, "flag provided")
	if !ok {
		t.Fatalf("stderr should indicate unknown/invalid flag for --dep; got %q", resp.Stderr)
	}
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
}
```
