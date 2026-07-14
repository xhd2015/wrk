## Expected

- Exit code is non-zero.
- Stdout is empty.
- Stderr contains `is not a git repository`.

## Errors

- cwd does not resolve to a git toplevel.

## Exit Code

- Non-zero

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
is not a git repository
</contains>`)
}
```
