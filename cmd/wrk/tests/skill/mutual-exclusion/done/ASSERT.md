## Expected

- Non-zero exit code.
- Stderr mentions mutual exclusion (or unexpected arguments).
- Stdout is empty.

## Errors

- `skill` cannot be combined with `--done`.

## Exit Code

- Non-zero

```go
import (
	"strings"

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
	lower := strings.ToLower(resp.Stderr)
	if strings.Contains(lower, "mutually exclusive") {
		return
	}
	if strings.Contains(lower, "unexpected") {
		return
	}
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}
```
