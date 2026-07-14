## Expected

- Exit code 0.
- Stdout empty (no path print).
- Stderr contains a notice that cwd is already the main repository root, including the main path.
- Fake interactive shell was **not** launched.

## Side Effects

- No nested shell; no follow-up file write.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
already at main repository root
</contains>`)
	wantMain := resolvePath(t, req.MainRepo)
	assertContains(t, resp.Stderr, wantMain)
	assertFakeShellNotLaunched(t, req)
}
```
