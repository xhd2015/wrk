## Expected

- Non-zero exit.
- Stderr mentions `--json` and the primary `--done`.
- Preferred implementer wording (soft): `wrk: --json is not valid with --done` (or equivalent policy that names both).
- Must not be a silent accept into merge-back.

## Errors

- `--json` cannot be combined with `--done`.

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for --done --json, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--json") {
		t.Fatalf("stderr should mention --json, got %q", se)
	}
	if !strings.Contains(se, "--done") {
		t.Fatalf("stderr should mention --done (primary-aware reject), got %q", se)
	}
	if !strings.Contains(se, "not valid") &&
		!strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "cannot") &&
		!strings.Contains(se, "only valid") {
		t.Fatalf("stderr should indicate reject policy for --json with --done, got %q", se)
	}
}
```
