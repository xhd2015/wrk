## Expected

- Non-zero exit.
- Flags are accepted (not "unrecognized flag").
- Stderr mentions mutual exclusion of terminal mode flags.

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
		t.Fatalf("expected non-zero for mutual terminal flags, stdout=%q", resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "unrecognized flag") {
		t.Fatalf("expected validation error after flags are implemented; got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "mutually exclusive") || strings.Contains(resp.Stderr, "exclusive") {
		return
	}
	assert.Output(t, resp.Stderr, `<contains>
mutually exclusive
</contains>`)
}
```
