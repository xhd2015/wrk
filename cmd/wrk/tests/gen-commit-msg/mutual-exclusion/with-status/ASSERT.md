## Expected

- Non-zero exit code.
- Stderr mentions mutual exclusion (and preferably `--gen-commit-msg` or `--status`).
- Stdout is empty.

## Errors

- `--gen-commit-msg` cannot be combined with `--status`.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}
```
