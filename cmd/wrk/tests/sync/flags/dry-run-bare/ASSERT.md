## Expected

- Non-zero exit code.
- Stderr contains the locked dry-run host list including `--sync` and `--propagate-tags`:
  `--dry-run is only valid with` (full host list may include --reinstall-local/--push/--gen-commit-msg)
- Stdout empty.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, resp.Stderr, "--dry-run is only valid with")
	// End-state: --all-deps is not a dry-run host (RED until implementer removes it from the list).
	if strings.Contains(resp.Stderr, "--all-deps") {
		t.Fatalf("stderr host list must not include --all-deps; got %q", resp.Stderr)
	}
	assertContains(t, resp.Stderr, "--sync")
	assertContains(t, resp.Stderr, "--propagate-tags")
}
```
