## Expected

- Non-zero exit code.
- Stderr clearly rejects the old subcommand (unknown action, expected flag, or
  similar).
- Stdout is empty (no accidental list success).

## Errors

- Subcommand form `skill list` is not accepted after the skill-cli flag migration.

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
		t.Fatalf("expected non-zero exit for old subcommand, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout should be empty for rejected subcommand, got %q", resp.Stdout)
	}
	if strings.TrimSpace(resp.Stderr) == "" {
		t.Fatal("expected clear stderr error for old skill subcommand")
	}
	lower := strings.ToLower(resp.Stderr)
	// Accept any clear rejection of the positional form (unknown / expected flag / usage).
	ok := strings.Contains(lower, "unknown") ||
		strings.Contains(lower, "expected") ||
		strings.Contains(lower, "unrecognized") ||
		strings.Contains(lower, "invalid") ||
		strings.Contains(lower, "usage") ||
		strings.Contains(lower, "--list") ||
		strings.Contains(lower, "flag") ||
		strings.Contains(lower, "subcommand")
	if !ok {
		t.Fatalf("stderr should clearly reject old subcommand, got %q", resp.Stderr)
	}
}
```
