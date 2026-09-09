## Expected

- Exit code 0.
- Stdout is empty (cwd `./myrepo` is a plain dir with no nested git repos).
- Stderr is empty.
- Saved project is not used (would have printed a status block for saved/myrepo).

## Side Effects

- `./myrepo` exists in cwd but is not a git repository; fallback must not run.

## Exit Code

- 0

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty (no fallback to saved), got %q", resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	// Guard: saved subject must not appear (would indicate projects.json fallback).
	if strings.Contains(resp.Stdout, "Dir:") {
		t.Fatalf("unexpected status block (fallback ran?): %q", resp.Stdout)
	}
	_ = req
}
```