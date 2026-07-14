## Expected

- Non-zero exit.
- Follow-up file empty / unchanged (no land).
- Stderr mentions mutual exclusion or both `--force-cd` and `--no-cd`.
- Stdout empty or free of a successful worktree path create (prefer empty).

## Exit Code

- Non-zero

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertFollowupEmpty(t, resp)
	// Prefer specific wording; fall back to both flag names.
	errText := resp.Stderr + "\n" + resp.Stdout
	hasForce := strings.Contains(errText, "--force-cd")
	hasNo := strings.Contains(errText, "--no-cd")
	hasMutual := strings.Contains(errText, "mutually exclusive") ||
		strings.Contains(errText, "mutual") ||
		strings.Contains(errText, "cannot be used together") ||
		strings.Contains(errText, "cannot combine")
	if hasMutual && (hasForce || hasNo) {
		return
	}
	if hasForce && hasNo {
		return
	}
	t.Fatalf("expected mutual-exclusion error mentioning --force-cd and/or --no-cd; stderr=%q stdout=%q",
		resp.Stderr, resp.Stdout)
}
```
