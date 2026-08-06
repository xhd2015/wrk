## Expected Output

```
would: git push --force-with-lease origin main
```

## Expected

- Exit code 0.
- Stdout is exactly the force-with-lease dry-run plan line (no `pushed …`).
- Stderr empty.
- Origin `refs/heads/main` unchanged from pre-run snapshot.
- Plan must include `--force-with-lease` (not bare `--force`).

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStderr(t, resp.Stderr)

	assert.Output(t, resp.Stdout, v2StdoutTemplate(wouldForcePushBranchLine("origin", "main")))
	if strings.Contains(resp.Stdout, "pushed ") {
		t.Fatalf("dry-run must not print pushed confirmation; got %q", resp.Stdout)
	}
	// Guard against bare --force plan (product must use --force-with-lease).
	if strings.Contains(resp.Stdout, "git push --force origin") &&
		!strings.Contains(resp.Stdout, "--force-with-lease") {
		t.Fatalf("plan must use --force-with-lease, not bare --force; got %q", resp.Stdout)
	}

	before := readOriginMainBefore(t, req)
	after := revParseRef(t, req.OriginBare, "refs/heads/main")
	if after != before {
		t.Fatalf("origin/main mutated under --dry-run: before %s after %s", before, after)
	}
}
```
