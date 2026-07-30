## Expected

- Non-zero exit; empty stdout.
- Stderr contains `unexpected arguments` (arity with `--main` is 0 positionals).
- Must **not** only claim mutual exclusion of `--main` with `--cd` (compose is valid without path).

## Errors

- Extra path after `--main --cd`.

## Exit Code

- Non-zero

```go
import (
	"strings"
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
unexpected arguments
</contains>`)
	// RED contract: path-with-compose is arity error, not exclusive-modes.
	// Soft: if implementer still only says mutually exclusive without unexpected args, fail.
	if !strings.Contains(resp.Stderr, "unexpected") {
		t.Fatalf("stderr should report unexpected arguments for extra path, got %q", resp.Stderr)
	}
}
```
