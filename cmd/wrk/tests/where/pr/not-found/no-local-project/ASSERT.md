## Expected

- Non-zero exit.
- Stdout empty.
- Stderr indicates no local project / no matching repo for the PR (prefer naming
  `acme/app` or “local project” / “no local”).

## Errors

- Location lookup finds no main with origin matching owner/repo.

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
		t.Fatalf("expected non-zero when no matching local project; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	se := strings.ToLower(resp.Stderr)
	// Prefer naming lack of local project and/or owner/repo.
	if !strings.Contains(se, "local") && !strings.Contains(se, "project") &&
		!strings.Contains(se, "acme/app") && !strings.Contains(se, "no ") {
		t.Fatalf("stderr should indicate no local matching project; got %q", resp.Stderr)
	}
}
```
