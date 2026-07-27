## Expected

- Flag layer accepts `--merge-back` + `--propagate-tags` (no mutual-exclusion error).
- Stderr must **not** contain `mutually exclusive`.

## Side Effects

- None required (flag-layer only).

## Exit Code

- Any (0 or non-zero), as long as the failure is not flag mutual exclusion.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer still rejects --merge-back --propagate-tags as mutually exclusive; stderr=%q stdout=%q exit=%d",
			se, resp.Stdout, resp.ExitCode)
	}
}
```