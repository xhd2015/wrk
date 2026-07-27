## Expected

- Non-zero exit code.
- Stderr reports `does not exist` for the cwd-resolved `saved/myrepo` path.
- Stdout is empty.

## Errors

- Path with separator is not eligible for basename fallback.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
does not exist
</contains>`)

	candidate := resolvePath(t, filepath.Join(req.RepoDir, "saved", "myrepo"))
	assertContains(t, resp.Stderr, candidate)
}
```