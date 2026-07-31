## Expected

- Non-zero exit code.
- Stderr mentions `--exec` and that only a **single** / one `--bring` is allowed (or multi not valid with exec).
- Prefer exact-ish: `wrk: --exec is only valid with a single --bring path`.
- Must not pass on bare `unrecognized flag: --exec` alone.
- No `external/` directory created (reject before materializing preferred).

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for multi --bring + --exec, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	se := resp.Stderr
	if !strings.Contains(se, "--exec") {
		t.Fatalf("stderr should mention --exec, got %q", se)
	}
	// Reject false-GREEN: only "unrecognized flag: --exec" without multi policy.
	if strings.Contains(se, "unrecognized flag") && !strings.Contains(se, "bring") {
		t.Fatalf("stderr is parse-unknown only (%q); want multi-bring + --exec policy error", se)
	}
	lower := strings.ToLower(se)
	ok := strings.Contains(lower, "single") ||
		strings.Contains(lower, "one ") ||
		strings.Contains(lower, "only valid") ||
		strings.Contains(lower, "multiple") ||
		strings.Contains(lower, "more than one") ||
		(strings.Contains(lower, "not valid") && strings.Contains(se, "--bring"))
	if !ok {
		t.Fatalf("stderr should say --exec needs a single --bring (or multi not allowed); got %q", se)
	}

	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
}
```
