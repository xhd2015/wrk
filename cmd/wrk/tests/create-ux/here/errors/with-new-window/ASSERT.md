## Expected

- Non-zero exit.
- Stderr mentions window / `--no-new-window` mutual exclusion (or window requires terminal).
- Not an unrecognized-flag error.

## Exit Code

- non-zero

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --here --new-window, stdout=%q", resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected validation error; got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "mutually exclusive") ||
		(strings.Contains(resp.Stderr, "window") && strings.Contains(resp.Stderr, "terminal")) {
		return
	}
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}
```
