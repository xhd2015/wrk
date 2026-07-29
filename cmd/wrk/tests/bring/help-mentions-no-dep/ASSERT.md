## Expected

- Exit code 0.
- Help text contains `--no-dep`.
- Help documents that `--no-dep` is worktree-only / skips replace+tidy (soft wording).
- Help for `-v` / `--verbose` mentions `go mod tidy` (or tidy logging).

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--no-dep") {
		t.Fatalf("help must mention --no-dep; got %q", help)
	}
	// Soft: worktree-only / skip tidy / no replace wording near --no-dep.
	if !bringHelpMentionsNoDepSemantics(help) {
		t.Fatalf("help for --no-dep should mention worktree-only / skip tidy (or replace); got %q", help)
	}
	// -v help should mention go mod tidy logging.
	if !bringHelpMentionsVerboseTidy(help) {
		t.Fatalf("help for -v/--verbose should mention go mod tidy; got %q", help)
	}
}

func bringHelpMentionsNoDepSemantics(help string) bool {
	for _, line := range strings.Split(help, "\n") {
		if !strings.Contains(line, "--no-dep") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "worktree") ||
			strings.Contains(lower, "tidy") ||
			strings.Contains(lower, "replace") ||
			strings.Contains(lower, "skip") {
			return true
		}
	}
	// Whole-help soft fallback if multi-line description.
	lower := strings.ToLower(help)
	return strings.Contains(lower, "--no-dep") &&
		(strings.Contains(lower, "skip") || strings.Contains(lower, "worktree only") ||
			strings.Contains(lower, "without replace") || strings.Contains(lower, "no replace"))
}

func bringHelpMentionsVerboseTidy(help string) bool {
	// Prefer the verbose help line itself.
	for _, line := range strings.Split(help, "\n") {
		if !strings.Contains(line, "--verbose") && !strings.Contains(line, "-v,") && !strings.Contains(line, "-v ") {
			// also match "  -v, --verbose"
			if !strings.Contains(line, "-v") {
				continue
			}
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "tidy") || strings.Contains(lower, "go mod") {
			return true
		}
	}
	lower := strings.ToLower(help)
	return strings.Contains(lower, "mod tidy") ||
		(strings.Contains(lower, "verbose") && strings.Contains(lower, "tidy"))
}
```
