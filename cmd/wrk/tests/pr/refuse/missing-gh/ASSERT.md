## Expected

- Non-zero exit.
- Stderr explains `--pr` requires GitHub CLI `gh`, with install URL `cli.github.com` preferred:
  e.g. `Error: wrk: --pr requires the GitHub CLI (gh); install from https://cli.github.com/`
- No PR success tokens on stdout.

## Errors

- Missing `gh` is a hard refuse with install guidance.

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
	se := resp.Stderr
	if !strings.Contains(se, "cli.github.com") &&
		!strings.Contains(strings.ToLower(se), "github cli") &&
		!strings.Contains(se, "requires") {
		t.Fatalf("stderr should install-prompt for gh (cli.github.com / GitHub CLI); got %q", se)
	}
	// Prefer mentioning gh and --pr.
	if !strings.Contains(se, "gh") {
		t.Fatalf("stderr should mention gh, got %q", se)
	}
	if strings.Contains(resp.Stdout, "PR created") || strings.Contains(resp.Stdout, "comment added") {
		t.Fatalf("must not claim PR success without gh; stdout=%q", resp.Stdout)
	}
}
```
