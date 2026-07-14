## Expected

- Non-zero exit code.
- Stderr is exactly `wrk: --where requires a path argument` plus trailing `\n`.
- Stdout is empty.

## Errors

- Missing required `--where` basename argument.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	// Hard-error stderr must end with trailing newline (main.go print path).
	assert.Output(t, resp.Stderr, "wrk: --where requires a path argument\n")
}
```