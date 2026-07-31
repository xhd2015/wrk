## Expected

- Non-zero exit.
- Stderr mentions `--comment` (empty after trim / required).
- No PR create via fake gh.

## Errors

- Comment must be non-empty after trim.

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
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for empty comment; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "--comment") {
		t.Fatalf("stderr should mention --comment, got %q", resp.Stderr)
	}
	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
}
```
