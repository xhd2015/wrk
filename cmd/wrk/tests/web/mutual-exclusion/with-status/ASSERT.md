## Expected

- Non-zero exit code.
- Stderr mentions mutual exclusion (and preferably `--web`).
- Stdout is empty.

## Errors

- `--web` cannot be combined with `--status`.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for --web --status, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}
```
