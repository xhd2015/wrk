## Expected

- Non-zero exit.
- Stderr indicates `--exec` is **not valid with `--list`** (mode policy), not merely that `--exec` is an unknown flag.
- Preferred implementer wording (soft): `wrk: --exec is not valid with --list` (or equivalent mutual-exclusion phrasing that names both flags).
- Must **not** pass on `unrecognized flag: --exec` alone.
- Stdout empty (no successful worktree list dump).

## Errors

- `--exec` cannot be combined with `--list`.

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
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}

	se := resp.Stderr
	// Reject false-GREEN: current binary only says "unrecognized flag: --exec".
	if strings.Contains(se, "unrecognized") {
		t.Fatalf("stderr is parse-unknown only (%q); want mode policy rejecting --exec with --list", se)
	}
	if !strings.Contains(se, "--exec") {
		t.Fatalf("stderr should mention --exec, got %q", se)
	}
	if !strings.Contains(se, "--list") {
		t.Fatalf("stderr should mention --list (mutual exclusion with --exec), got %q", se)
	}
	// Policy phrasing: not valid / mutually exclusive / cannot / only valid …
	if !strings.Contains(se, "not valid") &&
		!strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "cannot") &&
		!strings.Contains(se, "only valid") {
		t.Fatalf("stderr should indicate not-valid/mutual-exclusion policy, got %q", se)
	}
}
```
