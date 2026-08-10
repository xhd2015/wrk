## Expected

- Non-zero exit.
- Stderr indicates no tag / no version tag found.
- Consumer go.mod unchanged (replace still present).

## Errors

- No matching git tags under dep.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertGoModUnchanged(t, req)
	assertNotContains(t, resp.Stdout, "dep-update ")
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "tag") && !strings.Contains(se, "version") {
		t.Fatalf("stderr should indicate no tags, got %q", resp.Stderr)
	}
}
```
