## Expected

- Non-zero exit.
- Stderr mentions `--title` (required / missing).
- No PR success tokens; `gh pr create` not called.

## Errors

- Both `--title` and `--comment` are always required with `--pr`.

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
		t.Fatalf("expected non-zero without --title; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "--title") {
		t.Fatalf("stderr should mention --title, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "PR created") {
		t.Fatalf("must not create PR; stdout=%q", resp.Stdout)
	}
	invocs := parseFakeGhLog(t, ghLogPath(req))
	assertGhSubcmdNotCalled(t, invocs, "create")
}
```
