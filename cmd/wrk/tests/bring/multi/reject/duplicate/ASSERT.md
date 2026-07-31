## Expected

- Non-zero exit code.
- Stderr indicates duplicate / same path / already specified (soft class); should mention path or `--bring`.
- Prefer no `external/` directory (validate before create).
- Stdout empty preferred.

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
		t.Fatalf("expected non-zero for duplicate --bring paths, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	se := strings.ToLower(resp.Stderr)
	ok := strings.Contains(se, "duplicate") ||
		strings.Contains(se, "same") ||
		strings.Contains(se, "already") ||
		strings.Contains(se, "twice") ||
		strings.Contains(se, "repeated")
	if !ok {
		t.Fatalf("stderr should indicate duplicate bring path; got %q", resp.Stderr)
	}

	// Prefer fail before materializing.
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
}
```
