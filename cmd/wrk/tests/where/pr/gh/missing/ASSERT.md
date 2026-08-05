## Expected

- Non-zero exit.
- Stderr explains need for GitHub CLI `gh`, with install URL `cli.github.com` preferred.
- Stdout empty (no path printed).

## Errors

- Missing `gh` is a hard refuse with install guidance (same family as bare `--pr`).

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
		t.Fatalf("expected non-zero when gh missing; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	se := resp.Stderr
	if !strings.Contains(se, "cli.github.com") &&
		!strings.Contains(strings.ToLower(se), "github cli") &&
		!strings.Contains(se, "requires") {
		t.Fatalf("stderr should install-prompt for gh; got %q", se)
	}
	if !strings.Contains(se, "gh") {
		t.Fatalf("stderr should mention gh, got %q", se)
	}
}
```
